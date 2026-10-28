package gocloakbrowser

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	ErrBinaryNotFound                = errors.New("binary not found at expected path")
	ErrManifestFetchFailed           = errors.New("cannot fetch manifest")
	ErrSignatureFetchFailed          = errors.New("cannot fetch signature")
	ErrInvalidManifestSignature      = errors.New("invalid manifest signature")
	ErrChecksumMismatch              = errors.New("sha256 checksum mismatch")
	ErrBinaryNotFoundAfterExtraction = errors.New("binary not found after extraction")
)

// EnsureBinary checks if the binary is already available locally or downloads it.
func EnsureBinary(ctx context.Context, version string) (string, error) {
	// 1. Check local environment override
	if local := os.Getenv("CLOAKBROWSER_BINARY_PATH"); local != "" {
		if _, err := os.Stat(local); err == nil {
			return local, nil
		}
		return "", fmt.Errorf("CLOAKBROWSER_BINARY_PATH points to non-existent file: %s", local)
	}

	if version == "" {
		if envVer := os.Getenv("CLOAKBROWSER_VERSION"); envVer != "" {
			version = envVer
		} else {
			version = GetChromiumVersion()
		}
	}

	// 2. Check cached binary
	execPath, err := GetBinaryExecutablePath(version)
	if err != nil {
		return "", err
	}

	if info, err := os.Stat(execPath); err == nil && !info.IsDir() {
		// Found in cache
		slog.Debug("Found cached CloakBrowser binary", "version", version, "path", execPath)
		return execPath, nil
	}

	// 3. Download and extract
	tag, err := GetPlatformTag()
	if err != nil {
		return "", err
	}

	slog.Info("Downloading CloakBrowser binary", "version", version, "platform", tag)
	primaryURL, fallbackURL := GetDownloadURLs(version, tag)
	archiveData, err := downloadWithFallback(ctx, primaryURL, fallbackURL)
	if err != nil {
		return "", fmt.Errorf("download binary archive: %w", err)
	}

	// Verify SHA256 / Ed25519 if manifest is available
	manifestURL := fmt.Sprintf("%s/chromium-v%s/SHA256SUMS", "https://cloakbrowser.dev", version)
	sigURL := fmt.Sprintf("%s/chromium-v%s/SHA256SUMS.sig", "https://cloakbrowser.dev", version)
	if err := verifyArchiveIntegrity(ctx, archiveData, GetArchiveName(tag), manifestURL, sigURL); err != nil {
		slog.Debug("Archive verification skipped or failed", "error", err)
	}

	destDir, err := GetBinaryDir(version)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("create binary directory: %w", err)
	}

	if runtime.GOOS == "windows" {
		err = extractZip(archiveData, destDir)
	} else {
		err = extractTarGz(archiveData, destDir)
	}
	if err != nil {
		return "", fmt.Errorf("extract archive: %w", err)
	}

	// Mark executable
	_ = os.Chmod(execPath, 0755)

	if _, err := os.Stat(execPath); err != nil {
		return "", fmt.Errorf("%w: %s", ErrBinaryNotFoundAfterExtraction, execPath)
	}

	slog.Info("CloakBrowser binary extracted successfully", "path", execPath)
	return execPath, nil
}

func downloadWithFallback(ctx context.Context, primary, fallback string) ([]byte, error) {
	hc := &http.Client{Timeout: 10 * time.Minute}

	fetch := func(u string) ([]byte, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		resp, err := hc.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, u)
		}
		return io.ReadAll(resp.Body)
	}

	data, err := fetch(primary)
	if err == nil {
		return data, nil
	}

	return fetch(fallback)
}

func verifyArchiveIntegrity(ctx context.Context, archiveBytes []byte, filename, manifestURL, sigURL string) error {
	hc := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return err
	}
	resp, err := hc.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrManifestFetchFailed
	}
	manifest, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	reqSig, err := http.NewRequestWithContext(ctx, http.MethodGet, sigURL, nil)
	if err != nil {
		return err
	}
	respSig, err := hc.Do(reqSig)
	if err != nil || respSig.StatusCode != http.StatusOK {
		return ErrSignatureFetchFailed
	}
	sig, _ := io.ReadAll(respSig.Body)
	_ = respSig.Body.Close()

	// Verify Ed25519 signature of manifest against known public keys
	validSig := false
	for _, pubKeyB64 := range BinarySigningPubkeys {
		pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyB64)
		if err == nil && len(pubKeyBytes) == ed25519.PublicKeySize {
			if ed25519.Verify(pubKeyBytes, manifest, sig) {
				validSig = true
				break
			}
		}
	}
	if !validSig {
		return ErrInvalidManifestSignature
	}

	// Verify SHA256 of downloaded archive
	sum := sha256.Sum256(archiveBytes)
	hashHex := hex.EncodeToString(sum[:])

	lines := strings.Split(string(manifest), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && filepath.Base(parts[1]) == filename {
			if parts[0] != hashHex {
				return fmt.Errorf("%w: expected %s, got %s", ErrChecksumMismatch, parts[0], hashHex)
			}
			return nil
		}
	}

	return nil
}

func extractTarGz(archiveData []byte, targetDir string) error {
	gzReader, err := gzip.NewReader(bytes.NewReader(archiveData))
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(targetDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				_ = outFile.Close()
				return err
			}
			_ = outFile.Close()
		case tar.TypeSymlink:
			_ = os.Remove(target)
			_ = os.Symlink(header.Linkname, target)
		}
	}
	return nil
}

func extractZip(archiveData []byte, targetDir string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
	if err != nil {
		return err
	}

	for _, file := range zipReader.File {
		target := filepath.Join(targetDir, file.Name)
		if file.FileInfo().IsDir() {
			_ = os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			_ = rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		_ = rc.Close()
		_ = outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
