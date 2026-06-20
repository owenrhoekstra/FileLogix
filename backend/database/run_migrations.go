package database

import (
	"database/sql"
	"embed"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"

	"FileLogix/utilities/logger"
)

//go:embed migrations/*.sql
var migrations embed.FS

func RunMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "migration driver error: %v", err)
		log.Fatal(err)
	}

	d, err := iofs.New(migrations, "migrations")
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "migration source error: %v", err)
		log.Fatal(err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "migration init error: %v", err)
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Errorf(uuid.Nil, uuid.Nil, "migration failed: %v", err)
		log.Fatal(err)
	}

	logger.Infof(uuid.Nil, uuid.Nil, "database migrations applied successfully")
}
