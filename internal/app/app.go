package app

import (
	"log"
	"net/http"

	"github.com/mkukarin01/snort/internal/config"
	"github.com/mkukarin01/snort/internal/router"
	"github.com/mkukarin01/snort/internal/storage"
)

func Run() {
	cfg := config.NewConfig()

	cfg.Print()

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	r := router.NewRouter(cfg, store)
	log.Printf("Starting server on http://%s\n", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
