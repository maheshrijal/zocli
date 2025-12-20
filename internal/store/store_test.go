package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maheshrijal/zocli/internal/zomato"
)

func TestStore_SaveLoad(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "zocli_store_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "orders.json")
	s, err := New(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	orders := []zomato.Order{
		{ID: "1", Restaurant: "A", PlacedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
		{ID: "2", Restaurant: "B", PlacedAt: time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC)},
	}

	if err := s.Save(orders); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Store sorts by PlacedAt descending (latest first)
	// So loaded[0] should be ID "2"
	if len(loaded) != 2 {
		t.Fatalf("Loaded %d orders, want 2", len(loaded))
	}

	if loaded[0].ID != "2" {
		t.Errorf("First order ID = %s, want 2 (latest)", loaded[0].ID)
	}
	if loaded[1].ID != "1" {
		t.Errorf("Second order ID = %s, want 1", loaded[1].ID)
	}
}

func TestStore_LoadCorrupt(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zocli_store_corrupt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "orders.json")
	if err := os.WriteFile(dbPath, []byte("{invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}

	s, _ := New(dbPath)
	_, err = s.Load()
	if err == nil {
		t.Error("Load corrupt file: succeeded, want error")
	}
}

func TestDefaultPath(t *testing.T) {
	// This test just ensures no panic, as it depends on user config dir
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath failed: %v", err)
	}
	if path == "" {
		t.Error("DefaultPath returned empty string")
	}
}
