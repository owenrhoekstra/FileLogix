package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"FileLogix/utilities/logger"
)

const backupInterval = 24 * time.Hour

var backupDir = "/srv/FileLogix/backups"

func Start(ctx context.Context) {
	logger.Infof(uuid.Nil, uuid.Nil, "backup daemon started")

	safeRun(ctx)

	ticker := time.NewTicker(backupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Infof(uuid.Nil, uuid.Nil, "backup daemon shutting down")
			return
		case <-ticker.C:
			safeRun(ctx)
		}
	}
}

func run(ctx context.Context) error {
	logger.Infof(uuid.Nil, uuid.Nil, "backup daemon starting postgres dump")

	if err := dumpPostgres(ctx); err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "postgres dump: %w", err)
		return nil
	}

	if err := uploadToBackblaze(ctx); err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "backblaze upload: %w", err)
		return nil
	}

	if err := pruneOldDumps(); err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "failed to prune old backups: %w", err)
	}

	logger.Infof(uuid.Nil, uuid.Nil, "backup daemon completed successfully")
	return nil
}

func dumpPostgres(ctx context.Context) error {
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05Z")
	outPath := filepath.Join(backupDir, fmt.Sprintf("postgres_%s.dump", timestamp))

	cmd := exec.CommandContext(ctx, "pg_dump",
		"--format=custom",
		"--file", outPath,
		"--dbname", dbURL(),
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", os.Getenv("DB_PASSWORD")))

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pg_dump failed: %w — output: %s", err, string(out))
	}

	logger.Infof(uuid.Nil, uuid.Nil, "backup daemon postgres dump written to %s", outPath)
	return nil
}

func uploadToBackblaze(ctx context.Context) error {
	remote := os.Getenv("RCLONE_REMOTE") // e.g. "b2crypt:filelogix-backups"
	if remote == "" {
		logger.Errorf(uuid.Nil, uuid.Nil, "RCLONE_REMOTE not set")
		return nil
	}

	cmd := exec.CommandContext(ctx, "rclone", "copy", backupDir, remote)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("rclone copy failed: %w — output: %s", err, string(out))
	}

	logger.Infof(uuid.Nil, uuid.Nil, "backup daemon upload to backblaze complete")
	return nil
}

func dbURL() string {
	return fmt.Sprintf("postgres://%s:%s@/%s?host=/var/run/postgresql",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
}

func safeRun(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(uuid.Nil, uuid.Nil, "backup daemon panic: %v", r)
		}
	}()

	if err := run(ctx); err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "backup daemon run failed: %v", err)
	}
}

func pruneOldDumps() error {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -7)
	for _, e := range entries {
		info, _ := e.Info()
		if !e.IsDir() && info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(backupDir, e.Name()))
		}
	}
	return nil
}
