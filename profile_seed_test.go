package gocloakbrowser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cloak "github.com/nickecco/gocloakbrowser"
)

func TestProfileSeedPersistence(t *testing.T) {
	tmpDir := t.TempDir()

	args, err := cloak.PersistentSeedArgs(tmpDir, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(args) == 0 || !strings.HasPrefix(args[0], "--fingerprint=") {
		t.Fatalf("expected --fingerprint in args, got %v", args)
	}

	seedFile := filepath.Join(tmpDir, cloak.ProfileSeedFile)
	seed1, err := cloak.ReadProfileSeed(seedFile)
	if err != nil || seed1 < cloak.SeedMin || seed1 > cloak.SeedMax {
		t.Fatalf("expected valid seed in file, got %d (err: %v)", seed1, err)
	}

	// Second invocation must reuse the exact same seed
	args2, err := cloak.PersistentSeedArgs(tmpDir, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedArg := args[0]
	if args2[0] != expectedArg {
		t.Errorf("expected same seed arg %s on second launch, got %s", expectedArg, args2[0])
	}
}

func TestProfileSeedCorruptRecovery(t *testing.T) {
	tmpDir := t.TempDir()
	seedFile := filepath.Join(tmpDir, cloak.ProfileSeedFile)

	_ = os.WriteFile(seedFile, []byte("invalid-junk"), 0644)

	seed, err := cloak.ReadProfileSeed(seedFile)
	if err == nil || seed != 0 {
		t.Errorf("expected error reading corrupt seed, got seed %d", seed)
	}

	// File should be removed upon corruption detection
	if _, err := os.Stat(seedFile); !os.IsNotExist(err) {
		t.Errorf("corrupt seed file was not cleaned up")
	}
}
