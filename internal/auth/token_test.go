// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package auth

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestTokenData_IsExpired(t *testing.T) {
	past := &TokenData{Expiry: time.Now().Add(-time.Hour)}
	if !past.IsExpired() {
		t.Error("expected expired token to report IsExpired=true")
	}

	future := &TokenData{Expiry: time.Now().Add(time.Hour)}
	if future.IsExpired() {
		t.Error("expected fresh token to report IsExpired=false")
	}

	// Within the 10s safety margin is considered expired.
	margin := &TokenData{Expiry: time.Now().Add(5 * time.Second)}
	if !margin.IsExpired() {
		t.Error("expected token within 10s margin to report IsExpired=true")
	}
}

func TestSaveAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	td := &TokenData{
		AccessToken:  "access-abc",
		RefreshToken: "refresh-xyz",
		Expiry:       time.Now().Add(time.Hour).Truncate(time.Second),
		Email:        "alice@example.com",
		Groups:       []string{"tenant-acme"},
	}

	if err := Save(td); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.AccessToken != td.AccessToken {
		t.Errorf("AccessToken: got %q want %q", got.AccessToken, td.AccessToken)
	}
	if got.Email != td.Email {
		t.Errorf("Email: got %q want %q", got.Email, td.Email)
	}
	if len(got.Groups) != 1 || got.Groups[0] != "tenant-acme" {
		t.Errorf("Groups: got %v", got.Groups)
	}
}

func TestLoad_NotLoggedIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := Load()
	if !errors.Is(err, ErrNotLoggedIn) {
		t.Errorf("expected ErrNotLoggedIn, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := Save(&TokenData{AccessToken: "x", Expiry: time.Now().Add(time.Hour)}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := Delete(); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	path, _ := TokenPath()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected token file to be removed")
	}

	// Idempotent: second Delete should not error.
	if err := Delete(); err != nil {
		t.Errorf("second Delete: %v", err)
	}
}
