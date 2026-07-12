package rabbitmq

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"FileLogix/utilities/logger"

	"github.com/google/uuid"

	amqp "github.com/rabbitmq/amqp091-go"
)

var rabbitMQURL = fmt.Sprintf(
	"amqp://%s:%s@%s:%s/",
	os.Getenv("RABBITMQ_USER"),
	os.Getenv("RABBITMQ_PASSWORD"),
	os.Getenv("RABBITMQ_HOST"),
	os.Getenv("RABBITMQ_PORT"),
)

var (
	conn *amqp.Connection
	mu   sync.RWMutex
)

func Init() error {
	c, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "rabbitmq: initial dial failed: %v", err)
		return err
	}

	ch, err := c.Channel()
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "rabbitmq: initial channel open failed: %v", err)
		c.Close()
		return err
	}
	defer ch.Close()

	if err := declareTopology(ch); err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "rabbitmq: initial exchange declare failed: %v", err)
		c.Close()
		return err
	}

	mu.Lock()
	conn = c
	mu.Unlock()

	go handleReconnect(c)

	logger.Infof(uuid.Nil, uuid.Nil, "rabbitmq: connected")

	return nil
}

// Channel returns a fresh channel off the shared connection.
// Caller opens it, does work, closes it. This is your "per-transaction" unit.
func Channel() (*amqp.Channel, error) {
	mu.RLock()
	defer mu.RUnlock()
	if conn == nil {
		return nil, errors.New("rabbitmq: connection not initialized")
	}
	return conn.Channel()
}
