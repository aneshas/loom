package loom

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	DB DBConfig `yaml:"db"`

	Host string `yaml:"host"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

func MustLoadConfig[TConfig any](configPath string) *TConfig {
	config, err := LoadConfig[TConfig](configPath)
	if err != nil {
		panic(err)
	}

	return config
}

// LoadConfig loads configuration from a YAML file
// TODO - use viper so we can support flags and env vars
func LoadConfig[TConfig any](configPath string) (*TConfig, error) {
	// TODO - depending on the environment, we may want to support different config files
	configPath = path.Join(configPath, "dev.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config TConfig

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func (c *AppConfig) DBConn() (*sql.DB, error) {
	if c.IsSQLite() {
		return sql.Open("sqlite3", c.DB.Name+".db")
	}

	// TODO - add support for postgres
	// c.PostgresDSN()

	return nil, nil
}

func (c *AppConfig) IsSQLite() bool {
	return c.DB.Host == "" && c.DB.Port == 0 && c.DB.User == "" && c.DB.Password == ""
}

// SQLiteDSN returns the SQLite database connection string
func (c *AppConfig) SQLiteDSN() string {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("sqlite3://%s.db", c.DB.Name)
	}

	dbPath := filepath.Join(cwd, fmt.Sprintf("/%s.db", c.DB.Name))

	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return fmt.Sprintf("sqlite3://%s.db", c.DB.Name)
	}

	return fmt.Sprintf("sqlite3://%s", absPath)
}

// PostgresDSN returns the PostgreSQL database connection string
func (c *AppConfig) PostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DB.User, c.DB.Password, c.DB.Host, c.DB.Port, c.DB.Name)
}
