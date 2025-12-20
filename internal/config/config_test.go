package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_SaveLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zocli_config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfgPath := filepath.Join(tmpDir, "config.json")
	
	want := Config{Cookie: "secret-cookie"}
	if err := Save(cfgPath, want); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got.Cookie != want.Cookie {
		t.Errorf("Cookie = %q, want %q", got.Cookie, want.Cookie)
	}
}

func TestConfig_LoadMissing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zocli_config_test_missing")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfgPath := filepath.Join(tmpDir, "nonexistent.json")
	_, err = Load(cfgPath)
	if !os.IsNotExist(err) {
		t.Errorf("Load missing file err = %v, want os.IsNotExist", err)
	}
}

func TestConfig_PreferPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zocli_migration_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Case 1: Neither exists, prefer new
	newP := filepath.Join(tmpDir, "new", "config.json")
	oldP := filepath.Join(tmpDir, "old", "config.json")
	
	got := preferPath(newP, oldP, 0o600)
	if got != newP {
		t.Errorf("Case 1: got %s, want %s", got, newP)
	}

	// Case 2: Old exists, New missing -> Should migrate
	if err := os.MkdirAll(filepath.Dir(oldP), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldP, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	got = preferPath(newP, oldP, 0o600)
	if got != newP {
		t.Errorf("Case 2: got %s, want %s (should migrate)", got, newP)
	}
	if _, err := os.Stat(newP); os.IsNotExist(err) {
		t.Error("Case 2: Failed to migrate file to new path")
	}

	// Case 3: Both exist -> Prefer new
	got = preferPath(newP, oldP, 0o600)
	if got != newP {
		t.Errorf("Case 3: got %s, want %s", got, newP)
	}
}
