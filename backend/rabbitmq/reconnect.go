package rabbitmq

import (
	"time"

	"FileLogix/utilities/logger"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handleReconnect(c *amqp.Connection) {
	closeErr := <-c.NotifyClose(make(chan *amqp.Error))
	if closeErr == nil {
		// graceful close (e.g. we called conn.Close() ourselves on shutdown)
		return
	}

	logger.Errorf(uuid.Nil, uuid.Nil, "rabbitmq: connection closed unexpectedly: %v", closeErr)

	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		time.Sleep(backoff)

		newConn, err := amqp.Dial(rabbitMQURL)
		if err != nil {
			logger.Errorf(uuid.Nil, uuid.Nil, "rabbitmq: redial failed: %v", err)
			if backoff < maxBackoff {
				backoff *= 2
			}
			continue
		}

		ch, err := newConn.Channel()
		if err != nil {
			logger.Errorf(uuid.Nil, uuid.Nil, "rabbitmq: channel open failed after redial: %v", err)
			newConn.Close()
			if backoff < maxBackoff {
				backoff *= 2
			}
			continue
		}

		if err := declareTopology(ch); err != nil {
			logger.Errorf(uuid.Nil, uuid.Nil, "rabbitmq: exchange redeclare failed: %v", err)
			ch.Close()
			newConn.Close()
			if backoff < maxBackoff {
				backoff *= 2
			}
			continue
		}
		ch.Close()

		mu.Lock()
		conn = newConn
		mu.Unlock()

		logger.Infof(uuid.Nil, uuid.Nil, "rabbitmq: reconnected")

		go handleReconnect(newConn) // re-arm for the next drop
		return
	}
}
