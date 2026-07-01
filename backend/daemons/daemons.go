package daemons

import (
	"context"

	"FileLogix/daemons/backup"
	"FileLogix/daemons/healthcheck"
	"FileLogix/daemons/ocr"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func StartAll(ctx context.Context) {
	logger.Infof(uuid.Nil, uuid.Nil, "starting all daemons")
	go ocr.Start(ctx)
	go healthcheck.Start()
	go backup.Start(ctx)
}
