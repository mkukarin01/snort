package config

import (
   "flag"
   "fmt"
   "strings"
)

// Config структурка данных для конфига
type Config struct {
   Port       string
   BaseDomain string
   BasePath   string
   Address    string
   BaseURL    string
}

// NewConfig запускаем конфигурацию, наполняем структурку, данными из командой строки
func NewConfig() *Config {
   cfg := &Config{
      BaseDomain: "localhost",
   }

   flag.StringVar(&cfg.Port, "a", "8080", "Port for HTTP server")
   flag.StringVar(&cfg.BasePath, "b", "", "Base path for shortened links")

   flag.Parse()

   if (strings.HasPrefix(cfg.Port, "localhost:")) {
      after, found := strings.CutPrefix(cfg.Port, "localhost:")

      if (found) {
         cfg.Port = after
      }
   }

   cfg.Address = fmt.Sprintf("%s:%s", cfg.BaseDomain, cfg.Port)
   if cfg.BasePath != "" && !strings.HasPrefix(cfg.BasePath, "/") {
      cfg.BasePath = "/" + cfg.BasePath
   }
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
