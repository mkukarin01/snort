package main

import (
    "log"
    "net/http"

    "github.com/mkukarin01/snort/internal/app"
    "github.com/mkukarin01/snort/internal/config"
)

func main() {
    cfg := config.NewConfig()

    if err := cfg.Validate(); err != nil {
        log.Fatalf("Invalid configuration: %v", err)
    }

    r := app.NewRouter(cfg)
    log.Printf("Starting server on http://%s\n", cfg.Address)
    log.Fatal(http.ListenAndServe(cfg.Address, r))
}
