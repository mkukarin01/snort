package app

import (
	"log"
	"net/http"

	"github.com/mkukarin01/snort/internal/config"
	"github.com/mkukarin01/snort/internal/router"
)

func Run() {
	cfg := config.NewConfig()

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	r := router.NewRouter(cfg)
	log.Printf("Starting server on http://%s\n", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
