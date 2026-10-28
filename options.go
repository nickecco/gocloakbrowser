package gocloakbrowser

import (
	"github.com/nickecco/gocloakbrowser/human"
)

// Viewport defines dimensions for the browser page.
type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// LaunchOptions configures launching CloakBrowser with Rod.
type LaunchOptions struct {
	// BinaryPath is the path to the CloakBrowser Chromium executable.
	// If empty, gocloakbrowser checks CLOAKBROWSER_BINARY_PATH env var,
	// then local cache (~/.cloakbrowser), and downloads it if missing.
	BinaryPath string

	// BrowserVersion allows pinning an exact Chromium version.
	// Defaults to the platform's current supported version or CLOAKBROWSER_VERSION.
	BrowserVersion string

	// Headless runs the browser without a visible window. Default is false (headed).
	Headless bool

	// Proxy specifies the proxy URL (e.g. "http://user:pass@host:8080" or "socks5://host:1080").
	Proxy string

	// Args contains additional Chromium command-line flags.
	Args []string

	// StealthArgs enables default stealth arguments (--fingerprint=<seed>, platform spoofing, etc.).
	// Defaults to true when nil.
	StealthArgs *bool

	// Timezone sets the IANA timezone (e.g. "America/New_York") via --fingerprint-timezone.
	Timezone string

	// Locale sets the BCP 47 locale (e.g. "en-US") via --lang and --fingerprint-locale.
	Locale string

	// GeoIP auto-detects timezone, locale, and exit IP from proxy or egress IP.
	// Automatically sets --fingerprint-timezone, --fingerprint-locale, and --fingerprint-webrtc-ip.
	GeoIP bool

	// UserDataDir points to a directory for a persistent browser profile.
	// A stable seed is stored in <UserDataDir>/.cloakbrowser-seed so the profile
	// retains the same identity across restarts.
	UserDataDir string

	// ExtensionPaths contains filesystem paths to unpacked Chrome extensions.
	ExtensionPaths []string

	// Viewport overrides the default viewport size.
	// In headed mode, leaving this nil avoids emulating a fixed viewport,
	// allowing the page to match the real OS window.
	// In headless mode, leaving this nil defaults to 1920x947.
	Viewport *Viewport

	// Humanize enables human-like mouse curves, keyboard timing, and scroll patterns.
	Humanize bool

	// HumanPreset sets the humanize behavioral preset ("default" or "careful").
	HumanPreset human.Preset

	// HumanConfig allows fine-grained overrides for human-like behaviors.
	HumanConfig *human.Config
}

// IsStealthEnabled returns true unless StealthArgs was explicitly set to false.
func (o LaunchOptions) IsStealthEnabled() bool {
	if o.StealthArgs == nil {
		return true
	}
	return *o.StealthArgs
}
