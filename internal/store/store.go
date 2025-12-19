package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/maheshrijal/zocli/internal/zomato"
)

type Store struct {
	path string
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	newPath := filepath.Join(dir, "zocli", "orders.json")
	oldPath := filepath.Join(dir, "zomatocli", "orders.json")
	return preferPath(newPath, oldPath, 0o600), nil
}

func New(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("store path is required")
	}
	return &Store{path: path}, nil
}

func (s *Store) Load() ([]zomato.Order, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	var orders []zomato.Order
	if err := json.Unmarshal(data, &orders); err != nil {
		return nil, err
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].PlacedAt.After(orders[j].PlacedAt)
	})
	return orders, nil
}

func (s *Store) Save(orders []zomato.Order) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func preferPath(newPath, oldPath string, perm os.FileMode) string {
	if fileExists(newPath) {
		return newPath
	}
	if fileExists(oldPath) {
		if err := copyFile(oldPath, newPath, perm); err == nil {
			return newPath
		}
		return oldPath
	}
	return newPath
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func copyFile(src, dst string, perm os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, perm)
}
