package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jobscout/internal/config"
	"jobscout/internal/httpx"
	"jobscout/internal/llm"
	"jobscout/internal/store"
)

func main() {
	cfg := config.Load()

	st, err := store.New(cfg.DataDir)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	client := llm.New(cfg.AnthropicKey, cfg.Model)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpx.NewServer(cfg, st, client),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 90 * time.Second, // LLM calls can be slow
	}

	go func() {
		log.Printf("JobScout API on :%s (llm enabled: %v)", cfg.Port, client.Enabled())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
