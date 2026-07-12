package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueDefinition struct {
	Name       string
	RoutingKey string
}

// Add new queues here — declareTopology picks them up automatically,
// both on Init and on every reconnect.
var queues = []QueueDefinition{
	{Name: "ocr.pending", RoutingKey: "ocr.pending"},
}

func declareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(
		"filelogix.events",
		"topic",
		true,
		false, false, false,
		nil,
	); err != nil {
		return err
	}

	for _, q := range queues {
		if _, err := ch.QueueDeclare(
			q.Name,
			true, // durable
			false, false, false,
			nil,
		); err != nil {
			return err
		}

		if err := ch.QueueBind(
			q.Name,
			q.RoutingKey,
			"filelogix.events",
			false,
			nil,
		); err != nil {
			return err
		}
	}

	return nil
}
