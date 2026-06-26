// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TokenData holds the OIDC token stored on disk.
type TokenData struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	Email        string    `json:"email"`
	Groups       []string  `json:"groups"`
}

// IsExpired reports whether the access token has expired (with a 10s margin).
func (t *TokenData) IsExpired() bool {
	return time.Now().After(t.Expiry.Add(-10 * time.Second))
}

// TokenPath returns the path to the stored token file.
func TokenPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve config dir: %w", err)
	}
	return filepath.Join(dir, "smeltry", "token.json"), nil
}

// Load reads the token from disk. Returns ErrNotLoggedIn if the file is absent.
var ErrNotLoggedIn = errors.New("not logged in — run: smeltry auth login")

func Load() (*TokenData, error) {
	path, err := TokenPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotLoggedIn
		}
		return nil, fmt.Errorf("reading token: %w", err)
	}
	var td TokenData
	if err := json.Unmarshal(data, &td); err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	return &td, nil
}

// Save writes the token to disk (creating parent directories as needed).
func Save(td *TokenData) error {
	path, err := TokenPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := json.MarshalIndent(td, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling token: %w", err)
	}
	// O_TRUNC ensures the file is replaced atomically with 0600 permissions
	// even if a pre-existing file had broader permissions.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("opening token file: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("writing token file: %w", err)
	}
	return nil
}

// Delete removes the stored token from disk.
func Delete() error {
	path, err := TokenPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing token: %w", err)
	}
	return nil
}
