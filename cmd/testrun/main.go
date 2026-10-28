package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	cloak "github.com/nickecco/gocloakbrowser"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	binPath := os.Getenv("CLOAKBROWSER_BINARY_PATH")
	if binPath == "" {
		binPath = "/Users/nick/Desktop/SOFTS/w-anti-fraud/vendor/cloakbrowser/Chromium.app"
	}

	fmt.Printf("==> Launching CloakBrowser in HEADED mode: %s\n", binPath)

	browser, err := cloak.Launch(ctx, cloak.LaunchOptions{
		BinaryPath: binPath,
		Headless:   false, // Headed mode
		Humanize:   true,  // Human mouse/scroll
	})
	if err != nil {
		log.Fatalf("Failed to launch CloakBrowser: %v", err)
	}
	defer browser.Close()

	fmt.Println("==> Opening https://nowsecure.nl (Cloudflare Turnstile)...")
	page, err := browser.NewPage("https://nowsecure.nl")
	if err != nil {
		log.Fatalf("Failed to open page: %v", err)
	}

	page.MustWaitLoad()
	fmt.Printf("==> Page Title: %s\n", page.MustInfo().Title)

	// Wait for Turnstile check to complete
	for i := range 15 {
		time.Sleep(1 * time.Second)
		title := page.MustInfo().Title
		fmt.Printf("   [%ds] Title: %s\n", i+1, title)
		if title != "Just a moment..." {
			fmt.Println("==> Cloudflare Turnstile BYPASSED!")
			break
		}
	}

	fmt.Println("\n==> Окно открыто. Держим 5 секунд перед закрытием...")
	time.Sleep(5 * time.Second)

	fmt.Println("==> Done.")
}
