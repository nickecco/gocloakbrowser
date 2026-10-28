package gocloakbrowser

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Bundled and platform-specific Chromium versions.
const (
	DefaultChromiumVersion = "146.0.7680.177.5"
	ProfileSeedFile        = ".cloakbrowser-seed"
	SeedMin                = 10000
	SeedMax                = 99999
)

var PlatformChromiumVersions = map[string]string{
	"linux-x64":    "146.0.7680.177.5",
	"linux-arm64":  "146.0.7680.177.3",
	"darwin-arm64": "145.0.7632.109.2",
	"darwin-x64":   "145.0.7632.109.2",
	"windows-x64":  "146.0.7680.177.5",
}

// BinarySigningPubkeys are Ed25519 public keys (base64-encoded) for verifying downloaded binaries.
var BinarySigningPubkeys = []string{
	"MKFKwIhUcKWq5xTuNA0Ovg99njcDEcEJvmWYYhApvaU=",
}

// IgnoreDefaultArgs contains automation flags that leak bot signals and must be suppressed.
var IgnoreDefaultArgs = []string{
	"--enable-automation",
	"--enable-unsafe-swiftshader",
}

// DefaultViewport models a maximized Chrome on 1080p Windows: 1920x947 (screen minus ~85px UI).
var DefaultViewport = Viewport{
	Width:  1920,
	Height: 947,
}

// GetPlatformTag returns the CloakBrowser platform identifier (e.g. "darwin-arm64", "linux-x64").
func GetPlatformTag() (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	switch osName {
	case "darwin":
		switch arch {
		case "arm64":
			return "darwin-arm64", nil
		case "amd64":
			return "darwin-x64", nil
		}
	case "linux":
		switch arch {
		case "amd64":
			return "linux-x64", nil
		case "arm64":
			return "linux-arm64", nil
		}
	case "windows":
		switch arch {
		case "amd64":
			return "windows-x64", nil
		}
	}

	return "", fmt.Errorf("unsupported platform: %s/%s", osName, arch)
}

// GetChromiumVersion returns the target Chromium version for the current platform.
func GetChromiumVersion() string {
	tag, err := GetPlatformTag()
	if err == nil {
		if v, ok := PlatformChromiumVersions[tag]; ok {
			return v
		}
	}
	return DefaultChromiumVersion
}

// GetCacheDir returns the directory where CloakBrowser downloads and caches files.
// Defaults to ~/.cloakbrowser, overridable via CLOAKBROWSER_CACHE_DIR.
func GetCacheDir() (string, error) {
	if custom := os.Getenv("CLOAKBROWSER_CACHE_DIR"); custom != "" {
		return custom, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home dir: %w", err)
	}
	return filepath.Join(home, ".cloakbrowser"), nil
}

// GetBinaryDir returns the directory path for a specific Chromium version.
func GetBinaryDir(version string) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	if version == "" {
		version = GetChromiumVersion()
	}
	return filepath.Join(cacheDir, fmt.Sprintf("chromium-%s", version)), nil
}

// GetBinaryExecutablePath returns the path to the executable inside the binary directory.
func GetBinaryExecutablePath(version string) (string, error) {
	binDir, err := GetBinaryDir(version)
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(binDir, "Chromium.app", "Contents", "MacOS", "Chromium"), nil
	case "windows":
		return filepath.Join(binDir, "chrome.exe"), nil
	default:
		return filepath.Join(binDir, "chrome"), nil
	}
}

// GetArchiveName returns the filename of the download archive for the platform tag.
func GetArchiveName(tag string) string {
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("cloakbrowser-%s%s", tag, ext)
}

// GetDownloadURLs returns primary and fallback download URLs for the binary archive.
func GetDownloadURLs(version string, tag string) (primary string, fallback string) {
	baseURL := os.Getenv("CLOAKBROWSER_DOWNLOAD_URL")
	if baseURL == "" {
		baseURL = "https://cloakbrowser.dev"
	}
	archive := GetArchiveName(tag)
	primary = fmt.Sprintf("%s/chromium-v%s/%s", baseURL, version, archive)
	fallback = fmt.Sprintf("https://github.com/CloakHQ/cloakbrowser/releases/download/chromium-v%s/%s", version, archive)
	return primary, fallback
}
