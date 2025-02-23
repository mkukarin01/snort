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

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	db, err := storage.NewDatabase(cfg.DatabaseDSN)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}

	if db != nil {
		defer db.Close()
	}

	r := router.NewRouter(cfg, db)
	log.Printf("Starting server on http://%s\n", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}
