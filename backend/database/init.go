package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"FileLogix/utilities/logger"
)

var DB *sql.DB

func Init() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbName == "" {
		dbName = "filelogix"
	}

	connStr := fmt.Sprintf(
		"postgres://%s:%s@/%s?host=/var/run/postgresql&sslmode=disable",
		dbUser,
		dbPassword,
		dbName,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "failed to open database connection: %v", err)
		log.Fatal(err)
	}

	if err := DB.Ping(); err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "failed to ping database: %v", err)
		log.Fatal(err)
	}

	logger.Infof(uuid.Nil, uuid.Nil, "connected to PostgreSQL database: %s", dbName)
}
