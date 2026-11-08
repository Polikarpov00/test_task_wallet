package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Polikarpov00/test_task_wallet/internal/config"
	"github.com/Polikarpov00/test_task_wallet/internal/handler"
	"github.com/Polikarpov00/test_task_wallet/internal/repository/postgres"
	"github.com/Polikarpov00/test_task_wallet/internal/service"
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

	repo := postgres.NewWalletRepo(pool)
	svc := service.NewWalletService(repo)
	h := handler.New(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h.Register(mux)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Println("server started")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
