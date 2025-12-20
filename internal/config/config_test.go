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
