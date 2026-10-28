package gocloakbrowser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

var (
	ErrExitIPDiscoveryFailed = errors.New("failed to discover exit IP via echo services")
)

var ipEchoURLs = []string{
	"https://api.ipify.org",
	"https://checkip.amazonaws.com",
	"https://ifconfig.me/ip",
}

// CountryLocaleMap maps ISO country codes to their predominant BCP 47 locales.
var CountryLocaleMap = map[string]string{
	"US": "en-US", "GB": "en-GB", "AU": "en-AU", "CA": "en-CA", "NZ": "en-NZ",
	"IE": "en-IE", "ZA": "en-ZA", "SG": "en-SG",
	"DE": "de-DE", "AT": "de-AT", "CH": "de-CH",
	"FR": "fr-FR", "BE": "fr-BE",
	"ES": "es-ES", "MX": "es-MX", "AR": "es-AR", "CO": "es-CO", "CL": "es-CL",
	"BR": "pt-BR", "PT": "pt-PT",
	"IT": "it-IT", "NL": "nl-NL",
	"JP": "ja-JP", "KR": "ko-KR", "CN": "zh-CN", "TW": "zh-TW", "HK": "zh-HK",
	"RU": "ru-RU", "UA": "uk-UA", "PL": "pl-PL", "CZ": "cs-CZ", "RO": "ro-RO",
	"IL": "he-IL", "TR": "tr-TR", "SA": "ar-SA", "AE": "ar-AE", "EG": "ar-EG",
	"IN": "hi-IN", "ID": "id-ID", "PH": "en-PH",
	"TH": "th-TH", "VN": "vi-VN", "MY": "ms-MY",
	"SE": "sv-SE", "NO": "nb-NO", "DK": "da-DK", "FI": "fi-FI",
	"GR": "el-GR", "HU": "hu-HU", "BG": "bg-BG",
}

// GeoResult contains resolved IP, Timezone, and Locale.
type GeoResult struct {
	IP       string
	Timezone string
	Locale   string
}

// ResolveExitIP discovers the egress IP (optionally through a proxy).
func ResolveExitIP(ctx context.Context, proxyURLStr string) (string, error) {
	hc := &http.Client{
		Timeout: 10 * time.Second,
	}

	if proxyURLStr != "" {
		pURL, err := url.Parse(proxyURLStr)
		if err != nil {
			return "", fmt.Errorf("invalid proxy URL: %w", err)
		}
		hc.Transport = &http.Transport{
			Proxy: http.ProxyURL(pURL),
		}
	}

	for _, echoURL := range ipEchoURLs {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, echoURL, nil)
		if err != nil {
			continue
		}

		resp, err := hc.Do(req)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			continue
		}

		rawIP := strings.TrimSpace(string(body))
		addr, err := netip.ParseAddr(rawIP)
		if err == nil && addr.IsValid() {
			return addr.String(), nil
		}
	}

	return "", ErrExitIPDiscoveryFailed
}

// ResolveGeoLocation queries IP geolocation details (timezone, country).
func ResolveGeoLocation(ctx context.Context, proxyURLStr string) (*GeoResult, error) {
	ip, err := ResolveExitIP(ctx, proxyURLStr)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://ipwho.is/%s", ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return &GeoResult{IP: ip}, nil
	}

	hc := &http.Client{Timeout: 10 * time.Second}
	resp, err := hc.Do(req)
	if err != nil {
		slog.Debug("GeoIP lookup request failed", "error", err, "ip", ip)
		return &GeoResult{IP: ip}, nil
	}
	defer resp.Body.Close()

	var data struct {
		Success  bool   `json:"success"`
		Country  string `json:"country_code"`
		Timezone struct {
			ID string `json:"id"`
		} `json:"timezone"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || !data.Success {
		slog.Debug("GeoIP decode failed or response unsuccessful", "ip", ip)
		return &GeoResult{IP: ip}, nil
	}

	locale := "en-US"
	if loc, ok := CountryLocaleMap[data.Country]; ok {
		locale = loc
	}

	return &GeoResult{
		IP:       ip,
		Timezone: data.Timezone.ID,
		Locale:   locale,
	}, nil
}
