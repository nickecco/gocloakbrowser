package gocloakbrowser

import (
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

// DefaultStealthArgs returns the base stealth flags for the current platform.
func DefaultStealthArgs() []string {
	seed := RandomSeed()
	base := []string{
		"--no-sandbox",
		fmt.Sprintf("--fingerprint=%d", seed),
	}

	if runtime.GOOS == "darwin" {
		// Native Mac browser spoofing to avoid GPU/font mismatches.
		return append(base, "--fingerprint-platform=macos")
	}

	// Linux/Windows: Windows profile.
	return append(base, "--fingerprint-platform=windows")
}

// BuildArgs constructs deduplicated Chromium CLI arguments from stealth defaults and options.
func BuildArgs(opts LaunchOptions) []string {
	seen := make(map[string]string, 24)
	var order []string

	setArg := func(flag string) {
		key := strings.SplitN(flag, "=", 2)[0]
		if _, exists := seen[key]; !exists {
			order = append(order, key)
		}
		seen[key] = flag
	}

	// 1. Default stealth arguments
	if opts.IsStealthEnabled() {
		for _, arg := range DefaultStealthArgs() {
			setArg(arg)
		}
	}

	// 2. GPU blocklist bypass
	// Headed on all platforms (helps in Docker/Xvfb) or Windows (WebGPU Basic Render Driver)
	if !opts.Headless || runtime.GOOS == "windows" {
		setArg("--ignore-gpu-blocklist")
	}

	// 3. User-provided args
	for _, arg := range opts.Args {
		setArg(arg)
	}

	// 4. Font metrics profile adjustments
	if _, hasFont := seen["--fingerprint-windows-font-metrics"]; hasFont {
		currentVal := ""
		if feat, ok := seen["--enable-features"]; ok {
			parts := strings.SplitN(feat, "=", 2)
			if len(parts) == 2 {
				currentVal = parts[1]
			}
		}

		features := strings.Split(currentVal, ",")
		var cleanFeatures []string
		for _, f := range features {
			trimmed := strings.TrimSpace(f)
			if trimmed != "" {
				cleanFeatures = append(cleanFeatures, trimmed)
			}
		}

		if !slices.Contains(cleanFeatures, "MediaRouter") {
			cleanFeatures = append(cleanFeatures, "MediaRouter")
			setArg(fmt.Sprintf("--enable-features=%s", strings.Join(cleanFeatures, ",")))
		}
	}

	// 5. Timezone & Locale
	if opts.Timezone != "" {
		setArg(fmt.Sprintf("--fingerprint-timezone=%s", opts.Timezone))
	}
	if opts.Locale != "" {
		setArg(fmt.Sprintf("--lang=%s", opts.Locale))
		setArg(fmt.Sprintf("--fingerprint-locale=%s", opts.Locale))
	}

	// 6. Extensions
	if len(opts.ExtensionPaths) > 0 {
		var absPaths []string
		for _, p := range opts.ExtensionPaths {
			abs, err := filepath.Abs(p)
			if err == nil {
				absPaths = append(absPaths, abs)
			} else {
				absPaths = append(absPaths, p)
			}
		}
		joined := strings.Join(absPaths, ",")
		setArg(fmt.Sprintf("--load-extension=%s", joined))
		setArg(fmt.Sprintf("--disable-extensions-except=%s", joined))
	}

	// 7. Maximize window in headed mode
	hasWindowFlag := false
	windowFlags := []string{"--start-maximized", "--window-size", "--window-position"}
	for _, k := range windowFlags {
		if _, ok := seen[k]; ok {
			hasWindowFlag = true
			break
		}
	}

	if !opts.Headless && opts.Viewport == nil && !hasWindowFlag {
		setArg("--start-maximized")
	}

	// Build result keeping insertion order
	result := make([]string, 0, len(order))
	for _, key := range order {
		result = append(result, seen[key])
	}
	return result
}
