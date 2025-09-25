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
		Use:   "loom",
		Short: "Loom is an opinionated MVC web framework for Go — fast, simple, structured.",
		Long:  `Loom is a fast, convention-over-configuration web framework for Go — built to get you from zero to working, deployable app with minimal setup and maximum clarity. It enforces structure, favors simplicity, and helps you ship faster without sacrificing maintainability.`,
	}

	newCmd := &cobra.Command{
		Use:   "new [APP_NAME]",
		Short: "Generate a new loom web application",
		Long: `Generate a new loom web application with the specified name.
This command will create a new directory with a complete web application structure,
including all necessary files and configurations for a modern web app.

The new application will be created in the current directory.

Example:
  loom new myapp
  loom new myapp --module github.com/user/myapp
  loom new my-awesome-app`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appName := args[0]
			moduleName, _ := cmd.Flags().GetString("module")

			if err := runNewCommand(appName, moduleName); err != nil {
				fmt.Printf("Error creating new application: %v\n", err)
				os.Exit(1)
			}
		},
	}

	newCmd.Flags().StringP("module", "m", "", "Go module name (default: app name)")

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
  loom db migrate
  loom db migrate [dev|production]`,
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
  loom db seed
  loom db seed [dev|production]`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := os.MkdirAll("bin", 0o755); err != nil {
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

func runNewCommand(appName, moduleName string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Use app name as module name if not specified
	if moduleName == "" {
		moduleName = appName
	}

	fmt.Printf("Creating new Loom application '%s' with module '%s'...\n", appName, moduleName)

	// Generate the project using embedded templates
	if err := GenerateProject(appName, moduleName, cwd); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	projectPath := filepath.Join(cwd, appName)

	fmt.Printf("✓ Project structure created\n")

	//
	fmt.Printf("Running go mod replace ...\n") // only for dev
	modRepl := exec.Command("go", "mod", "edit", "-replace", "github.com/aneshas/loom=../")
	modRepl.Dir = projectPath
	modRepl.Stdout = os.Stdout
	modRepl.Stderr = os.Stderr

	if err := modRepl.Run(); err != nil {
		fmt.Printf("Warning: failed to run go mod replace: %v\n", err)
	} else {
		fmt.Printf("✓ Loom replaced\n")
	}
	//

	// Run go mod tidy to clean up dependencies
	fmt.Printf("Running go mod tidy...\n")
	modTidyCmd := exec.Command("go", "mod", "tidy")
	modTidyCmd.Dir = projectPath
	modTidyCmd.Stdout = os.Stdout
	modTidyCmd.Stderr = os.Stderr

	if err := modTidyCmd.Run(); err != nil {
		fmt.Printf("Warning: failed to run go mod tidy: %v\n", err)
	} else {
		fmt.Printf("✓ Dependencies resolved\n")
	}

	// Print success message and next steps
	fmt.Printf("\n🎉 Successfully created '%s'!\n\n", appName)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  cd %s\n", appName)
	fmt.Printf("  go run cmd/%s/main.go\n\n", appName)
	fmt.Printf("Your new Loom application will be available at http://localhost:8080\n")

	return nil
}
