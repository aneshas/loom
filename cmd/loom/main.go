package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aneshas/loom"
	"github.com/aneshas/loom/internal/db"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "goon",
		Short: "Goon is an opinionated MVC web framework for Go — fast, simple, structured.",
		Long:  `Goon is a fast, convention-over-configuration web framework for Go — built to get you from zero to working, deployable app with minimal setup and maximum clarity. It enforces structure, favors simplicity, and helps you ship faster without sacrificing maintainability.`,
	}

	newCmd := &cobra.Command{
		Use:   "new [APP_PATH]",
		Short: "Generate a new goon web application",
		Long: `Generate a new goon web application at the specified path.
This command will create a new directory with a complete web application structure,
including all necessary files and configurations for a modern web app.

Example:
  goon new myapp
  goon new ./myapp
  goon new /path/to/myapp`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appPath := args[0]
			fmt.Printf("new command called with APP_PATH: %s - web app generation will be implemented here\n", appPath)
		},
	}

	// DB command
	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Database operations",
		Long:  `Database operations including migrations and schema management.`,
	}

	migrateCmd := &cobra.Command{
		Use:   "migrate [CONFIG]",
		Short: "Run database migrations",
		Long: `Run all database migrations from ./internal/db/migrations directory.
The command will automatically detect whether to use SQLite or PostgreSQL based on the configuration.
If no CONFIG argument is provided, it defaults to 'dev'.

Example:
  goon db migrate
  goon db migrate [dev|production]`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// configName := "dev" - TODO

			// if len(args) > 0 {
			// 	configName = args[0]
			// }

			// TODO - Set env somehow and make load config use it

			cfg, err := loadConfig()
			if err != nil {
				fmt.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}

			// Check if migrations directory exists
			migrationsPath := "./internal/db/migrations"
			if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
				fmt.Printf("Error: Migrations directory not found: %s\n", migrationsPath)
				os.Exit(1)
			}

			// Convert to absolute path for migrate package
			absMigrationsPath, err := filepath.Abs(migrationsPath)
			if err != nil {
				fmt.Printf("Error getting absolute path: %v\n", err)
				os.Exit(1)
			}

			// Run migrations
			if err := db.RunMigrations(cfg, absMigrationsPath); err != nil {
				fmt.Printf("Error running migrations: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Database migrations completed successfully!")
		},
	}

	seedCmd := &cobra.Command{
		Use:   "seed",
		Short: "Run database seeders",
		Long: `Run all database seeders from ./scripts/seed.go file.

Example:
  goon db seed
  goon db seed [dev|production]`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := os.MkdirAll("bin", 0755); err != nil {
				fmt.Printf("Error creating bin directory: %v\n", err)
				os.Exit(1)
			}

			buildCmd := exec.Command("go", "build", "-o", "bin/seed", "./cmd/seed")
			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr

			if err := buildCmd.Run(); err != nil {
				fmt.Printf("Error compiling seed program: %v\n", err)
				os.Exit(1)
			}

			// TODO - sqlboiler -c sqlboiler.sqlite3.toml sqlite3
			// depending on the database type

			seedCmd := exec.Command("./bin/seed")
			seedCmd.Stdout = os.Stdout
			seedCmd.Stderr = os.Stderr

			if err := seedCmd.Run(); err != nil {
				fmt.Printf("Error running seed program: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Seeding completed successfully!")
		},
	}

	dbCmd.AddCommand(migrateCmd)
	dbCmd.AddCommand(seedCmd)

	rootCmd.AddCommand(newCmd, dbCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type cfg struct {
	loom.AppConfig `yaml:"app"`
}

func loadConfig() (*loom.AppConfig, error) {
	cfg, err := loom.LoadConfig[cfg]("./config")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	return &cfg.AppConfig, nil
}
