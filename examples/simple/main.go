package main

import (
	"context"
	"fmt"
	"log"
	"time"

	cloak "github.com/nickecco/gocloakbrowser"
)

func main() {
	ctx := context.Background()

	// Launch CloakBrowser backed by go-rod/rod
	browser, err := cloak.Launch(ctx, cloak.LaunchOptions{
		// Set to path of local Chromium build, or leave empty to auto-download / use cache:
		// BinaryPath: "/path/to/custom/chromium",
		Headless: false,
		Humanize: true, // Enables natural human-like mouse curves, delays, and typing
		// GeoIP: true,  // Auto-detect timezone/locale from proxy or current IP
	})
	if err != nil {
		log.Fatalf("Failed to launch CloakBrowser: %v", err)
	}
	defer browser.Close()

	// Create a new tab and navigate
	page, err := browser.NewPage("https://bot.incolumitas.com")
	if err != nil {
		log.Fatalf("Failed to open page: %v", err)
	}

	fmt.Printf("Page loaded: %s\n", page.MustInfo().Title)

	// Perform human-like scroll down
	time.Sleep(1 * time.Second)
	if err := page.HumanScroll(0, 400); err != nil {
		log.Printf("Scroll error: %v", err)
	}

	time.Sleep(2 * time.Second)
	fmt.Println("Completed test session successfully.")
}
