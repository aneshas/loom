package db

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode"

	"github.com/aneshas/loom"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/mattn/go-sqlite3"
)

func GenMigration(description string) error {
	fmt.Printf("Generating migration file with description: %s\n", description)

	description = strings.ToLower(description)
	description = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, description)
	description = strings.ReplaceAll(description, " ", "_")

	migrationsPath := "internal/db/migrations"

	// check if there are any files in internal/db/migrations
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(migrationsPath, 0755)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("failed to read migrations directory: %v", err)
		}
	}

	max := 0

	if len(files) > 0 {
		for _, file := range files {
			parts := strings.Split(file.Name(), "_")

			if len(parts) < 2 {
				continue
			}

			num, err := strconv.Atoi(parts[0])
			if err != nil {
				continue
			}

			if num > max {
				max = num
			}
		}
	}

	base := fmt.Sprintf("%05d_%s", max+1, description)
	up := fmt.Sprintf("%s.up.sql", base)
	down := fmt.Sprintf("%s.down.sql", base)

	err = os.WriteFile(path.Join(migrationsPath, up), []byte(""), 0644)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %v", err)
	}

	err = os.WriteFile(path.Join(migrationsPath, down), []byte(""), 0644)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %v", err)
	}

	return nil
}

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
