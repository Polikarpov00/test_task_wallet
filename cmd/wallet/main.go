package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Polikarpov00/test_task_wallet/internal/config"
	"github.com/Polikarpov00/test_task_wallet/internal/repository/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	err = postgres.Migrate(ctx, pool)
	if err != nil {
		log.Fatal(err)
	}

	port := cfg.HTTP.Port

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Println("server started")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
