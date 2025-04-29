package main

import (
	"fmt"

	"github.com/gavinmcnair/kubedb/internal/api"
	"github.com/gavinmcnair/kubedb/internal/store"
	"github.com/gavinmcnair/kubedb/internal/config"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize store
	db, err := store.NewBadgerDBStore(cfg.DBPath)
	if err != nil {
		fmt.Printf("Error initializing store: %v\n", err)
		return
	}
	defer db.Close()

	// Start API server
	api.StartServer(db, cfg)
}

