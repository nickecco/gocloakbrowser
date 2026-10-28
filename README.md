# gocloakbrowser

Stealth Chromium automation library for **Go (Golang)** backed by [go-rod/rod](https://github.com/go-rod/rod).

A 1-to-1 Go port of the official [CloakBrowser](https://cloakbrowser.dev) Python / JavaScript / .NET clients, designed to bypass anti-bot systems (Cloudflare Turnstile, reCAPTCHA, FingerprintJS, Kasada, BrowserScan) using a patched Chromium binary and behavioral humanization.

---

## ✨ Features

- **Native Go & Rod Integration**: Pure Go, no external Node.js dependencies, seamlessly embedded with `*rod.Browser` and `*rod.Page`.
- **Source-Level C++ Fingerprint Compatibility**: Prepares all required stealth flags (`--fingerprint=<seed>`, platform spoofing, WebRTC exit IP, timezone, locale).
- **Automation Signal Scrubbing**: Strips detectable automation flags (`--enable-automation`, `--enable-unsafe-swiftshader`) that expose `navigator.webdriver = true`.
- **Anti-Detect Viewport Logic**:
  - In **headed mode**: removes default emulated viewports (`devices.Clear`) so pages track the real OS window (`outerWidth >= innerWidth`).
  - In **headless mode**: applies a coherent, deterministic `1920x947` viewport.
- **Humanize Behavioral Layer (`humanize=true`)**:
  - **Mouse**: Bézier curves with minimum-jerk acceleration curves, Fitts' law duration planning, and target overshooting.
  - **Keyboard**: Natural keystroke delays, key dwell times, and flight times between characters.
  - **Scroll**: Smooth mouse wheel bursts with inertia and easing.
- **GeoIP & Egress Detection**: Auto-detects proxy egress IP via echo services, matching timezone, locale, and WebRTC spoofing.
- **Persistent Profile Identity**: Automatically stores and reuses `.cloakbrowser-seed` inside `UserDataDir`.
- **Automatic Binary Management**: Optionally auto-downloads and verifies the official CloakBrowser binary via Ed25519 signature checks, or accepts any local binary path.

---

## 📦 Installation

```bash
go get github.com/nickecco/gocloakbrowser
```

---

## 🚀 Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	cloak "github.com/nickecco/gocloakbrowser"
)

func main() {
	ctx := context.Background()

	// Launch CloakBrowser
	browser, err := cloak.Launch(ctx, cloak.LaunchOptions{
		// Set binary path or leave empty to auto-download / use CLOAKBROWSER_BINARY_PATH:
		// BinaryPath: "/path/to/cloakbrowser/chrome",
		Headless: false,
		Humanize: true, // Enable human-like mouse, keyboard, and scroll
		GeoIP:    true, // Match timezone, locale and WebRTC to proxy/egress IP
	})
	if err != nil {
		log.Fatalf("Failed to launch: %v", err)
	}
	defer browser.Close()

	// Open a new tab
	page, err := browser.NewPage("https://example.com")
	if err != nil {
		log.Fatalf("Failed to open page: %v", err)
	}

	// Humanized interactions
	err = page.HumanClick("button#submit")
	err = page.HumanInput("input#search", "hello world")
	err = page.HumanScroll(0, 500)

	// Direct access to all native rod methods:
	title := page.MustInfo().Title
	fmt.Printf("Page Title: %s\n", title)
}
```

---

## ⚙️ Launch Options

| Option | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `BinaryPath` | `string` | `""` | Path to CloakBrowser Chromium binary. Falls back to `CLOAKBROWSER_BINARY_PATH`, cache `~/.cloakbrowser`, or auto-download. |
| `BrowserVersion` | `string` | current | Specific version pin (e.g. `"146.0.7680.177.5"`). |
| `Headless` | `bool` | `false` | Run without a visible browser window. |
| `Proxy` | `string` | `""` | Proxy URL (e.g. `http://user:pass@host:port` or `socks5://host:port`). |
| `GeoIP` | `bool` | `false` | Discover egress IP and automatically configure timezone, locale, and WebRTC spoofing. |
| `Timezone` | `string` | `""` | IANA timezone (e.g. `America/New_York`). |
| `Locale` | `string` | `""` | BCP 47 locale (e.g. `en-US`). |
| `UserDataDir` | `string` | `""` | Directory for persistent profile. Stably maintains seed in `.cloakbrowser-seed`. |
| `ExtensionPaths`| `[]string`| `nil` | Paths to unpacked Chrome extensions. |
| `Humanize` | `bool` | `false` | Enable human-like mouse curves, delays, and typing. |
| `HumanPreset` | `Preset` | `"default"`| `"default"` or `"careful"`. |
| `Args` | `[]string` | `nil` | Additional Chromium command-line flags. |

---

## 🧠 Full Access to `rod`

`gocloakbrowser.Browser` embeds `*rod.Browser` and `gocloakbrowser.Page` embeds `*rod.Page`. You never lose any functionality provided by `rod`:

```go
// Direct rod usage
elements := page.MustElements("div.card")
page.MustScreenshot("page.png")
```

---

## 🧪 Testing

```bash
go test -v ./...
```

---

## 📄 License

MIT
