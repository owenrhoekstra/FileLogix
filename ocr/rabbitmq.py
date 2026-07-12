"""
Reusable RabbitMQ consumer for FileLogix Python services.

Copy this file into each service. Per-service code only needs to:
  1. Import `consumer` (or construct a new Consumer(...))
  2. Register handlers with @consumer.handler("routing.key")
  3. Call consumer.start() on FastAPI startup, consumer.stop() on shutdown

Design:
  - Runs pika's BlockingConnection in a background thread (see main chat
    discussion for why: no extra process/infra needed per service).
  - Exchange is "filelogix.events" (topic), matching the Go publisher.
  - Each service declares its OWN queue + binds it to the routing keys it
    cares about — queues belong to consumers, not publishers.
  - Retry-with-limit: on handler failure, message is requeued up to
    MAX_RETRIES times (tracked via a header FileLogix adds itself, since
    RabbitMQ's built-in x-death count needs a DLX setup we're not using yet).
    After MAX_RETRIES, the message is dropped (logged loudly) instead of
    requeued forever, to avoid poison-message infinite loops.
"""

import json
import logging
import os
import threading
import time
from typing import Callable

import pika
from pika.exchange_type import ExchangeType

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("rabbitmq")

RABBITMQ_USER = os.environ.get("RABBITMQ_USER", "guest")
RABBITMQ_PASSWORD = os.environ.get("RABBITMQ_PASSWORD", "guest")
RABBITMQ_HOST = os.environ.get("RABBITMQ_HOST", "rabbitmq")
RABBITMQ_PORT = os.environ.get("RABBITMQ_PORT", "5672")

RABBITMQ_URL = f"amqp://{RABBITMQ_USER}:{RABBITMQ_PASSWORD}@{RABBITMQ_HOST}:{RABBITMQ_PORT}/"
EXCHANGE_NAME = "filelogix.events"
MAX_RETRIES = 5
RETRY_HEADER = "x-filelogix-retry-count"


class Consumer:
    def __init__(self, queue_name: str, routing_keys: list[str]):
        self.queue_name = queue_name
        self.routing_keys = routing_keys
        self._handlers: dict[str, Callable[[dict], None]] = {}
        self._thread: threading.Thread | None = None
        self._stop_event = threading.Event()
        self._connection: pika.BlockingConnection | None = None

    def handler(self, routing_key: str):
        """Decorator to register a handler for a routing key.

        Usage:
            @consumer.handler("ocr.pending")
            def handle_ocr_pending(payload: dict):
                ...
        """
        def decorator(func: Callable[[dict], None]):
            self._handlers[routing_key] = func
            return func
        return decorator

    def start(self):
        self._thread = threading.Thread(target=self._run_with_reconnect, daemon=True)
        self._thread.start()

    def stop(self):
        self._stop_event.set()
        if self._connection and self._connection.is_open:
            try:
                self._connection.close()
            except Exception as e:
                logger.error("rabbitmq: error closing connection during stop: %s", e)
        if self._thread:
            self._thread.join(timeout=5)

    def _run_with_reconnect(self):
        backoff = 1
        max_backoff = 30

        while not self._stop_event.is_set():
            try:
                self._connect_and_consume()
                backoff = 1  # reset after a clean run
            except Exception as e:
                logger.error("rabbitmq: consumer loop error: %s", e)

            if self._stop_event.is_set():
                return

            logger.info("rabbitmq: reconnecting in %ss", backoff)
            time.sleep(backoff)
            backoff = min(backoff * 2, max_backoff)

    def _connect_and_consume(self):
        params = pika.URLParameters(RABBITMQ_URL)
        self._connection = pika.BlockingConnection(params)
        channel = self._connection.channel()

        channel.exchange_declare(
            exchange=EXCHANGE_NAME,
            exchange_type=ExchangeType.topic,
            durable=True,
        )

        channel.queue_declare(queue=self.queue_name, durable=True)

        for rk in self.routing_keys:
            channel.queue_bind(
                queue=self.queue_name,
                exchange=EXCHANGE_NAME,
                routing_key=rk,
            )

        channel.basic_qos(prefetch_count=1)
        channel.basic_consume(
            queue=self.queue_name,
            on_message_callback=self._on_message,
        )

        logger.info(
            "rabbitmq: consuming queue=%s routing_keys=%s",
            self.queue_name,
            self.routing_keys,
        )
        channel.start_consuming()

    def _on_message(self, channel, method, properties, body):
        routing_key = method.routing_key
        handler = self._handlers.get(routing_key)

        if handler is None:
            logger.error("rabbitmq: no handler registered for routing_key=%s", routing_key)
            channel.basic_ack(delivery_tag=method.delivery_tag)
            return

        try:
            payload = json.loads(body)
        except json.JSONDecodeError as e:
            logger.error("rabbitmq: failed to decode message body: %s", e)
            channel.basic_ack(delivery_tag=method.delivery_tag)  # malformed, not retryable
            return

        try:
            handler(payload)
            channel.basic_ack(delivery_tag=method.delivery_tag)
        except Exception as e:
            headers = (properties.headers or {}) if properties else {}
            retry_count = headers.get(RETRY_HEADER, 0)

            logger.error(
                "rabbitmq: handler failed for routing_key=%s (attempt %d): %s",
                routing_key,
                retry_count + 1,
                e,
            )

            if retry_count >= MAX_RETRIES:
                logger.error(
                    "rabbitmq: max retries exceeded for routing_key=%s, dropping message",
                    routing_key,
                )
                channel.basic_ack(delivery_tag=method.delivery_tag)
                return

            # nack without requeue, republish manually with incremented retry header
            # (simpler than relying on a DLX we haven't set up yet)
            channel.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
            new_headers = dict(headers)
            new_headers[RETRY_HEADER] = retry_count + 1

            channel.basic_publish(
                exchange=EXCHANGE_NAME,
                routing_key=routing_key,
                body=body,
                properties=pika.BasicProperties(
                    content_type="application/json",
                    delivery_mode=2,  # persistent
                    headers=new_headers,
                ),
            )