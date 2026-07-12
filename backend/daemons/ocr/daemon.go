package ocr

import (
	"context"
	"time"

	"github.com/google/uuid"

	"FileLogix/utilities/logger"
)

const pollInterval = 5 * time.Minute

func Start(ctx context.Context) {
	return
	logger.Infof(uuid.Nil, uuid.Nil, "ocr daemon started")

	if err := resetStuck(); err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "ocr daemon failed to reset stuck jobs: %v", err)
	}

	for {
		found := safeRun(ctx)

		if found {
			continue
		}

		select {
		case <-ctx.Done():
			logger.Infof(uuid.Nil, uuid.Nil, "ocr daemon shutting down")
			return
		case <-time.After(pollInterval):
		}
	}
}

func run(ctx context.Context) (bool, error) {
	file, err := claimNext()
	if err != nil {
		return false, err
	}
	if file == nil {
		return false, nil
	}

	logger.Infof(uuid.Nil, uuid.Nil, "ocr daemon processing file: %s", file.ID)

	text, err := callOCR(ctx, file.Path)
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "ocr failed for file %s: %v", file.ID, err)
		if repoErr := markFailed(file.ID); repoErr != nil {
			logger.Errorf(uuid.Nil, uuid.Nil, "failed to mark file %s as failed: %v", file.ID, repoErr)
		}
		return true, err
	}

	if err := saveResult(file.ID, file.DocumentID, text); err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "failed to save ocr result for file %s: %v", file.ID, err)
		if repoErr := markFailed(file.ID); repoErr != nil {
			logger.Errorf(uuid.Nil, uuid.Nil, "failed to mark file %s as failed: %v", file.ID, repoErr)
		}
		return true, err
	}

	logger.Infof(uuid.Nil, uuid.Nil, "ocr daemon completed file: %s", file.ID)
	return true, nil
}

func safeRun(ctx context.Context) (found bool) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(uuid.Nil, uuid.Nil, "ocr daemon panic: %v", r)
		}
	}()

	found, err := run(ctx)
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "ocr daemon run failed: %v", err)
	}

	return found
}
