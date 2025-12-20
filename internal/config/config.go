package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	Cookie string `json:"cookie"`
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	newPath := filepath.Join(dir, "zocli", "config.json")
	return newPath, nil
}

func Load(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if path == "" {
		return errors.New("config path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
