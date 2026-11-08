package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := postgres.Migrate(ctx, pool); err != nil {
		log.Fatal(err)
	}

	repo := postgres.NewWalletRepo(pool)
	svc := service.NewWalletService(repo)
	h := handler.New(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h.Register(mux)

	server := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	go func() {
		log.Printf("server started on :%s", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
