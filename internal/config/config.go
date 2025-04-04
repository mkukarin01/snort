package config

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

// Config структурка данных для конфига
type Config struct {
	Port            string
	BaseDomain      string
	BasePath        string
	Address         string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	SecretKey       string
}

// NewConfig запускаем конфигурацию, наполняем структурку, данными из командой строки
func NewConfig() *Config {
	cfg := &Config{
		BaseDomain: "localhost",
	}

	// переменные окружения
	envPort := os.Getenv("SERVER_ADDRESS")
	envBasePath := os.Getenv("BASE_URL")
	envFileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	envDatabaseDSN := os.Getenv("DATABASE_DSN")

	// аргументы/флаги/etc
	flag.StringVar(&cfg.Port, "a", "8080", "Port for HTTP server")
	flag.StringVar(&cfg.BasePath, "b", "", "Base path for shortened links")
	flag.StringVar(&cfg.FileStoragePath, "f", "./storage.json", "Path to file storage for shortened links")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database connection string (PostgreSQL)")

	flag.Parse()

	// если есть переменные окружения, они сверху
	if envPort != "" {
		cfg.Port = envPort
	}
	if envBasePath != "" {
		cfg.BasePath = envBasePath
	}
	if envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}
	if envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}
	// настроим секретный ключ, вдруг попросят его передавать
	cfg.SecretKey = "supersecretkey"

	// приводим порт к виду порта 8080 например
	if cfg.Port != "" && cfg.Port != "8080" {
		_, port, err := net.SplitHostPort(cfg.Port)

		if err != nil {
			fmt.Printf("Invalid address format for Port, fallback to default: %v\n", err)
			cfg.Port = "8080"
		}

		if port == "" {
			fmt.Println("Port is empty. fallback to default")
			cfg.Port = "8080"
		} else {
			cfg.Port = port
		}
	}

	// приводим basePath в порядок
	if cfg.BasePath != "" {
		parsedURL, err := url.Parse(cfg.BasePath)
		if err != nil {
			fmt.Printf("Invalid URL format for BasePath, fallback to default: %v\n", err)
			cfg.BasePath = "/"
		} else {
			cfg.BasePath = parsedURL.Path
		}
	}

	if cfg.BasePath != "" && !strings.HasPrefix(cfg.BasePath, "/") {
		cfg.BasePath = "/" + cfg.BasePath
	}

	// прост записали все в конфиг обратно
	cfg.Address = fmt.Sprintf("%s:%s", cfg.BaseDomain, cfg.Port)
	cfg.BaseURL = fmt.Sprintf("http://%s%s", cfg.Address, cfg.BasePath)

	return cfg
}

// Validate свалидируем конфиг
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	if c.BaseDomain == "" {
		return fmt.Errorf("base domain cannot be empty")
	}
	return nil
}
