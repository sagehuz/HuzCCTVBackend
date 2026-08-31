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
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/term"

	"huzbackend-go/internal/auth"
	"huzbackend-go/internal/cli"
	"huzbackend-go/internal/config"
	"huzbackend-go/internal/scan"
	"huzbackend-go/internal/signal"
	"huzbackend-go/internal/store"
	"huzbackend-go/internal/web"
)

//go:embed public
var publicFS embed.FS

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		// Run directly from a terminal → open the interactive menu (config, start/stop...).
		// If stdin is not a terminal (background script) → run the server as before.
		if stdinIsTTY() {
			chdirToExecDir()
			os.Exit(cli.RunMenu())
		}
		runServer()
		return
	}
	switch args[0] {
	case "serve":
		chdirToExecDir()
		runServer()
		return
	case "start", "stop", "restart", "status", "logs", "open",
		"autostart", "config", "menu", "help", "version", "--help", "-h":
		chdirToExecDir()
		os.Exit(cli.Run(args))
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q (run '%s help' for usage)\n",
			args[0], filepath.Base(os.Args[0]))
		os.Exit(2)
	}
}

// stdinIsTTY reports whether stdin is a terminal (interactive user).
// Used to decide between opening the menu and running the server when the
// binary is invoked without arguments. It checks via ioctl (x/term) so that
// /dev/null, pipes and regular files are correctly treated as non-terminal.
func stdinIsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// chdirToExecDir changes the working directory to the directory containing the
// binary so that relative paths (.env, data/, .huzbackend.pid, .huzbackend.log)
// are always resolved consistently, even when the server runs in the background
// or starts automatically at login.
func chdirToExecDir() {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("Warning: could not get executable path: %v", err)
		return
	}
	dir := filepath.Dir(exe)
	if err := os.Chdir(dir); err != nil {
		log.Printf("Warning: could not change working directory to %s: %v", dir, err)
	}
}

func runServer() {
	log.Printf("Huz CCTV Server %s (%s/%s) starting", cli.Version, runtime.GOOS, runtime.GOARCH)

	cfg := config.Load()

	storeSvc, err := store.NewStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer storeSvc.Close()

	authSvc, err := auth.NewAuthService(storeSvc, cfg)
	if err != nil {
		log.Fatalf("failed to initialize auth service: %v", err)
	}
	if err := authSvc.EnsureAdmin(); err != nil {
		log.Fatalf("failed to create default admin: %v", err)
	}

	scanner := scan.NewScanner()
	hub := signal.NewHub(func(token string) bool {
		_, ok := authSvc.ValidateToken(token)
		return ok
	})
	staticFS, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("failed to initialize static FS: %v", err)
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

	log.Printf("Server is running on port %s", cfg.Port)

	<-stopCh
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown HTTP error: %v", err)
	}

	hub.Shutdown()
	log.Println("Server stopped")
	if err := storeSvc.Close(); err != nil {
		log.Printf("close DB error: %v", err)
	}
	fmt.Println("Huz CCTV Server shutdown complete")
}
