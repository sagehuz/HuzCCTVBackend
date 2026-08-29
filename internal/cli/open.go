package cli

import (
	"fmt"
	"os"

	"huzbackend-go/internal/config"
)

func runOpen(args []string) int {
	cfg := config.Load()
	url := "http://127.0.0.1:" + cfg.Port
	if !portOpen(cfg.Port) {
		fmt.Fprintf(os.Stderr, "Warning: the server does not seem to be running on port %s — the page may not open.\n", cfg.Port)
	}
	fmt.Println("Opening:", url)
	if err := openBrowser(url); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open the browser: %v\n", err)
		fmt.Printf("Please open it manually: %s\n", url)
		return 1
	}
	return 0
}
