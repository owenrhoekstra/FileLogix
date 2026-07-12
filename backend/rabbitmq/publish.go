package rabbitmq

import (
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Publish(routingKey string, body []byte) error {
	ch, err := Channel()
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "rabbitmq: publish failed to open channel: %v", err)
		return err
	}
	defer ch.Close()

	if err := ch.Publish(
		"filelogix.events",
		routingKey,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	); err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "rabbitmq: publish to %s failed: %v", routingKey, err)
		return err
	}

	return nil
}
