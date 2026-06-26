// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package auth

import (
	"testing"
)

func TestSaveAndLoadConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg := &Config{
		IssuerURL: "https://auth.example.com/application/o/smeltry/",
		ClientID:  "smeltry-cli",
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	got, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got.IssuerURL != cfg.IssuerURL {
		t.Errorf("IssuerURL: got %q want %q", got.IssuerURL, cfg.IssuerURL)
	}
	if got.ClientID != cfg.ClientID {
		t.Errorf("ClientID: got %q want %q", got.ClientID, cfg.ClientID)
	}
}

func TestLoadConfig_Missing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error when config file does not exist")
	}
}
