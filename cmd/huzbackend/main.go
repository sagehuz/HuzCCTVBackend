package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	osSignal "os/signal"
	"syscall"
	"time"

	"huzbackend-go/internal/auth"
	"huzbackend-go/internal/config"
	"huzbackend-go/internal/scan"
	"huzbackend-go/internal/signal"
	"huzbackend-go/internal/store"
	"huzbackend-go/internal/web"
)

//go:embed public
var publicFS embed.FS

func main() {
	cfg := config.Load()

	storeSvc, err := store.NewStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("không thể khởi tạo DB: %v", err)
	}
	defer storeSvc.Close()

	authSvc, err := auth.NewAuthService(storeSvc, cfg)
	if err != nil {
		log.Fatalf("không thể khởi tạo auth: %v", err)
	}
	if err := authSvc.EnsureAdmin(); err != nil {
		log.Fatalf("không thể tạo admin mặc định: %v", err)
	}

	scanner := scan.NewScanner()
	hub := signal.NewHub()
	staticFS, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("không thể khởi tạo static FS: %v", err)
	}
	handler := web.NewHandler(cfg, authSvc, scanner, hub, staticFS)

	srv := &http.Server{
		Addr:              "0.0.0.0:" + cfg.Port,
		Handler:           handler.Mux(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stopCh := make(chan os.Signal, 1)
	osSignal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
			stopCh <- syscall.SIGTERM
		}
	}()

	log.Printf("Server đang chạy tại port %s", cfg.Port)

	<-stopCh
	log.Println("Đang đóng server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown HTTP error: %v", err)
	}

	hub.Shutdown()
	log.Println("Server đã dừng")
	if err := storeSvc.Close(); err != nil {
		log.Printf("close DB error: %v", err)
	}
	fmt.Println("Huz CCTV Server shutdown complete")
}
