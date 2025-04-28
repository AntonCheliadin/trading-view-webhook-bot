package database

import (
	"fmt"
	"go.uber.org/zap"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

func NewPostgresConnection() (*sqlx.DB, error) {

	connStr := os.Getenv("DATABASE_URL")

	if connStr == "" {
		connStr = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
		)
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	initMigrations(db)

	return db, nil
}

func initMigrations(db *sqlx.DB) {
	migrations := &migrate.FileMigrationSource{
		Dir: "./migrations",
	}

	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		zap.S().Errorf("Error during applying migrations! %s", err.Error())
	}
	zap.S().Infof("Applied %d migrations!", n)
}
