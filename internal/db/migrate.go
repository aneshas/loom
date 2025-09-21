package db

import (
	"fmt"
	"log"

	"github.com/aneshas/loom"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/mattn/go-sqlite3"
)

// RunMigrations executes all migrations from the specified directory
func RunMigrations(cfg *loom.AppConfig, migrationsPath string) error {
	var m *migrate.Migrate
	var err error

	if cfg.IsSQLite() {
		// Use SQLite
		log.Printf("Using SQLite database: %s", cfg.SQLiteDSN())

		// Create migrate instance with file source and SQLite database
		m, err = migrate.New(
			fmt.Sprintf("file://%s", migrationsPath),
			cfg.SQLiteDSN(),
		)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance: %w", err)
		}
	} else {
		// Use PostgreSQL
		log.Printf("Using PostgreSQL database: %s", cfg.PostgresDSN())

		// Create migrate instance with file source and PostgreSQL database
		m, err = migrate.New(
			fmt.Sprintf("file://%s", migrationsPath),
			cfg.PostgresDSN(),
		)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance: %w", err)
		}
	}

	defer func() {
		if m != nil {
			if _, err := m.Close(); err != nil {
				log.Printf("Error closing migration: %v", err)
			}
		}
	}()

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
