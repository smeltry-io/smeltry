// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the OIDC issuer and client settings persisted between logins.
type Config struct {
	IssuerURL string `json:"issuer_url"`
	ClientID  string `json:"client_id"`
}

// SaveConfig writes the OIDC config to disk with 0600 permissions.
func SaveConfig(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(cfg)
}

// LoadConfig reads the OIDC config from disk.
func LoadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("loading config (run 'smeltry auth login --issuer-url <url>'): %w", err)
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving config directory: %w", err)
	}
	return filepath.Join(dir, "smeltry", "config.json"), nil
}
