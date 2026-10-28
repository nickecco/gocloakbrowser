package gocloakbrowser

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/nickecco/gocloakbrowser/human"
)

// resolveExecutablePath checks if path points to a macOS .app bundle and resolves the binary.
func resolveExecutablePath(path string) string {
	if runtime.GOOS == "darwin" {
		cleaned := strings.TrimSuffix(path, "/")
		if strings.HasSuffix(cleaned, ".app") {
			candidate := filepath.Join(cleaned, "Contents", "MacOS", "Chromium")
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
		candidate := filepath.Join(path, "Contents", "MacOS", "Chromium")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return path
}

// Launch starts a CloakBrowser instance connected to Rod with stealth flags and settings.
func Launch(ctx context.Context, opts LaunchOptions) (*Browser, error) {
	// 1. Resolve binary path
	binPath := opts.BinaryPath
	if binPath == "" {
		resolved, err := EnsureBinary(ctx, opts.BrowserVersion)
		if err != nil {
			return nil, fmt.Errorf("resolve CloakBrowser binary: %w", err)
		}
		binPath = resolved
	}

	binPath = resolveExecutablePath(binPath)

	if _, err := os.Stat(binPath); err != nil {
		return nil, fmt.Errorf("cloakbrowser executable not found at %s: %w", binPath, err)
	}

	// 2. Resolve GeoIP if requested
	if opts.GeoIP {
		geo, err := ResolveGeoLocation(ctx, opts.Proxy)
		if err == nil && geo != nil {
			if opts.Timezone == "" && geo.Timezone != "" {
				opts.Timezone = geo.Timezone
			}
			if opts.Locale == "" && geo.Locale != "" {
				opts.Locale = geo.Locale
			}
			if geo.IP != "" {
				// Inject WebRTC exit IP spoofing
				opts.Args = append(opts.Args, fmt.Sprintf("--fingerprint-webrtc-ip=%s", geo.IP))
			}
		}
	}

	// 3. Persistent profile seed handling
	if opts.UserDataDir != "" {
		seededArgs, err := PersistentSeedArgs(opts.UserDataDir, opts.IsStealthEnabled(), opts.Args)
		if err != nil {
			return nil, fmt.Errorf("persistent profile seed: %w", err)
		}
		opts.Args = seededArgs
	}

	// 4. Build Chromium CLI args
	chromeArgs := BuildArgs(opts)

	// 5. Setup Rod launcher
	l := launcher.New().
		Bin(binPath).
		Headless(opts.Headless)

	// Suppress bot-tell automation flags
	for _, ign := range IgnoreDefaultArgs {
		flagName := strings.TrimPrefix(ign, "--")
		l.Delete(flags.Flag(flagName))
	}

	// Handle proxy
	var proxyUser, proxyPass string
	if opts.Proxy != "" {
		pURL, err := url.Parse(opts.Proxy)
		if err == nil && pURL.User != nil {
			proxyUser = pURL.User.Username()
			proxyPass, _ = pURL.User.Password()
			pURL.User = nil
			l.Proxy(pURL.String())
		} else {
			l.Proxy(opts.Proxy)
		}
	}

	// Set user data dir if specified
	if opts.UserDataDir != "" {
		l.UserDataDir(opts.UserDataDir)
	}

	// Apply built Chrome arguments to launcher
	for _, arg := range chromeArgs {
		trimmed := strings.TrimPrefix(arg, "--")
		parts := strings.SplitN(trimmed, "=", 2)
		flagName := flags.Flag(parts[0])
		if len(parts) == 2 {
			l.Set(flagName, parts[1])
		} else {
			l.Set(flagName)
		}
	}

	slog.Debug("Launching CloakBrowser",
		"headless", opts.Headless,
		"binary", binPath,
		"argsCount", len(chromeArgs),
	)

	// 6. Launch browser process
	wsURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch CloakBrowser process: %w", err)
	}

	// 7. Connect Rod client
	rodBrowser := rod.New().ControlURL(wsURL)
	if err := rodBrowser.Connect(); err != nil {
		l.Kill()
		return nil, fmt.Errorf("connect rod over cdp: %w", err)
	}

	// Setup proxy authentication handler if credentials were provided
	if proxyUser != "" {
		go rodBrowser.HandleAuth(proxyUser, proxyPass)()
	}

	// 8. Device / Viewport configuration
	// In headed mode: clear emulation so page tracks real OS window (avoid outerWidth < innerWidth tell).
	// In headless mode: apply coherent 1920x947 viewport unless overridden.
	if !opts.Headless {
		rodBrowser.DefaultDevice(devices.Clear)
	} else if opts.Viewport == nil {
		rodBrowser.DefaultDevice(devices.Device{
			Screen: devices.Screen{
				DevicePixelRatio: 1,
				Horizontal: devices.ScreenSize{
					Width:  DefaultViewport.Width,
					Height: DefaultViewport.Height,
				},
				Vertical: devices.ScreenSize{
					Width:  DefaultViewport.Width,
					Height: DefaultViewport.Height,
				},
			},
		})
	} else {
		rodBrowser.DefaultDevice(devices.Device{
			Screen: devices.Screen{
				DevicePixelRatio: 1,
				Horizontal: devices.ScreenSize{
					Width:  opts.Viewport.Width,
					Height: opts.Viewport.Height,
				},
				Vertical: devices.ScreenSize{
					Width:  opts.Viewport.Width,
					Height: opts.Viewport.Height,
				},
			},
		})
	}

	// Resolve humanize configuration
	humanCfg := human.ResolveConfig(opts.HumanPreset, opts.HumanConfig)

	return &Browser{
		Browser:     rodBrowser,
		launcher:    l,
		humanize:    opts.Humanize,
		humanConfig: humanCfg,
	}, nil
}
