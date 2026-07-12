package rabbitmq

import (
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func Close() {
	mu.RLock()
	c := conn
	mu.RUnlock()

	if c == nil {
		return
	}

	if err := c.Close(); err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "rabbitmq: error during close: %v", err)
		return
	}

	logger.Infof(uuid.Nil, uuid.Nil, "rabbitmq: connection closed gracefully")
}
