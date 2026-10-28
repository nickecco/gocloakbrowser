package gocloakbrowser_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	cloak "github.com/nickecco/gocloakbrowser"
)

func TestPrintDefaultLauncherArgs(t *testing.T) {
	l := launcher.New()
	t.Log("before delete has enable-automation:", l.Has(flags.Flag("enable-automation")))
	l.Delete(flags.Flag("enable-automation"))
	t.Log("after delete has enable-automation:", l.Has(flags.Flag("enable-automation")))
	for _, a := range l.FormatArgs() {
		if strings.Contains(a, "automation") {
			t.Errorf("found automation flag: %s", a)
		}
	}
}

func TestBuildArgsDefaultStealth(t *testing.T) {
	opts := cloak.LaunchOptions{
		Headless: true,
	}

	args := cloak.BuildArgs(opts)

	hasFingerprint := false
	hasPlatform := false
	for _, a := range args {
		if strings.HasPrefix(a, "--fingerprint=") {
			hasFingerprint = true
		}
		if runtime.GOOS == "darwin" && a == "--fingerprint-platform=macos" {
			hasPlatform = true
		} else if runtime.GOOS != "darwin" && a == "--fingerprint-platform=windows" {
			hasPlatform = true
		}
	}

	if !hasFingerprint {
		t.Errorf("expected --fingerprint flag, got args: %v", args)
	}
	if !hasPlatform {
		t.Errorf("expected platform spoofing flag for %s, got args: %v", runtime.GOOS, args)
	}
}

func TestBuildArgsCustomTimezoneAndLocale(t *testing.T) {
	opts := cloak.LaunchOptions{
		Timezone: "Europe/Berlin",
		Locale:   "de-DE",
	}

	args := cloak.BuildArgs(opts)

	hasTz := false
	hasLang := false
	hasFpLocale := false

	for _, a := range args {
		if a == "--fingerprint-timezone=Europe/Berlin" {
			hasTz = true
		}
		if a == "--lang=de-DE" {
			hasLang = true
		}
		if a == "--fingerprint-locale=de-DE" {
			hasFpLocale = true
		}
	}

	if !hasTz {
		t.Errorf("missing --fingerprint-timezone flag in %v", args)
	}
	if !hasLang {
		t.Errorf("missing --lang flag in %v", args)
	}
	if !hasFpLocale {
		t.Errorf("missing --fingerprint-locale flag in %v", args)
	}
}

func TestBuildArgsOverrideUserArgs(t *testing.T) {
	opts := cloak.LaunchOptions{
		Args: []string{"--fingerprint=54321"},
	}

	args := cloak.BuildArgs(opts)

	count := 0
	for _, a := range args {
		if a == "--fingerprint=54321" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("expected exactly one overridden --fingerprint=54321, got %d occurrences in %v", count, args)
	}
}

func TestBuildArgsDisabledStealth(t *testing.T) {
	disabled := false
	opts := cloak.LaunchOptions{
		StealthArgs: &disabled,
		Args:        []string{"--custom-flag"},
	}

	args := cloak.BuildArgs(opts)

	for _, a := range args {
		if strings.HasPrefix(a, "--fingerprint") {
			t.Errorf("did not expect fingerprint flag when stealth is disabled, got %s", a)
		}
	}
}
