package daemons

import (
	"context"

	"FileLogix/daemons/ocr"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func StartAll(ctx context.Context) {
	logger.Infof(uuid.Nil, uuid.Nil, "starting all daemons")
	go ocr.Start(ctx)
}
