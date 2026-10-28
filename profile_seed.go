package gocloakbrowser

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// RandomSeed generates a random 5-digit seed between 10000 and 99999.
func RandomSeed() int {
	return SeedMin + rand.IntN(SeedMax-SeedMin+1)
}

// ReadProfileSeed reads the seed from the specified file.
// Returns 0 if missing or corrupt.
func ReadProfileSeed(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	text := strings.TrimSpace(string(data))
	val, err := strconv.Atoi(text)
	if err != nil || val < SeedMin || val > SeedMax {
		_ = os.Remove(path)
		return 0, fmt.Errorf("corrupt seed in %s: %q", path, text)
	}
	return val, nil
}

// WriteProfileSeed writes the seed to the file atomically.
func WriteProfileSeed(path string, seed int) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp := fmt.Sprintf("%s.%d.tmp", path, rand.Uint64())
	content := fmt.Sprintf("%d\n", seed)

	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmp)
	}()

	// Atomic replace
	return os.Rename(tmp, path)
}

// PersistentSeedArgs checks if UserDataDir is set and stealthArgs are on.
// If so, it retrieves or creates the persistent seed in <userDataDir>/.cloakbrowser-seed
// and prepends `--fingerprint=<seed>` to args.
func PersistentSeedArgs(userDataDir string, stealthEnabled bool, args []string) ([]string, error) {
	if !stealthEnabled || userDataDir == "" {
		return args, nil
	}

	for _, a := range args {
		if strings.HasPrefix(a, "--fingerprint=") || a == "--fingerprint" {
			return args, nil
		}
	}

	seedPath := filepath.Join(userDataDir, ProfileSeedFile)
	seed, err := ReadProfileSeed(seedPath)
	if err != nil || seed == 0 {
		seed = RandomSeed()
		if writeErr := WriteProfileSeed(seedPath, seed); writeErr != nil {
			return args, fmt.Errorf("failed to write persistent seed: %w", writeErr)
		}
	}

	return append([]string{fmt.Sprintf("--fingerprint=%d", seed)}, args...), nil
}
