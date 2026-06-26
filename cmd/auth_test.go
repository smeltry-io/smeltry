// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/smeltry-io/smeltry/internal/auth"
)

// isolateConfig redirects ~/.config/smeltry to a temp dir for the duration of t.
func isolateConfig(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func TestAuthLogout_RemovesToken(t *testing.T) {
	isolateConfig(t)
	if err := auth.Save(&auth.TokenData{
		AccessToken: "tok",
		Expiry:      time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var buf bytes.Buffer
	cmd := newAuthLogoutCmd()
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if !strings.Contains(buf.String(), "Logged out") {
		t.Errorf("expected 'Logged out' in output, got: %q", buf.String())
	}

	// Token must no longer be loadable.
	if _, err := auth.Load(); err == nil {
		t.Error("expected ErrNotLoggedIn after logout, got nil")
	}
}

func TestAuthLogout_Idempotent(t *testing.T) {
	isolateConfig(t)
	// Logout without prior login must not error.
	cmd := newAuthLogoutCmd()
	cmd.SetOut(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Errorf("logout without prior login: %v", err)
	}
}

func TestAuthStatus_ValidToken(t *testing.T) {
	isolateConfig(t)
	if err := auth.Save(&auth.TokenData{
		AccessToken: "tok",
		Email:       "alice@example.com",
		Groups:      []string{"tenant-acme"},
		Expiry:      time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var buf bytes.Buffer
	cmd := newAuthStatusCmd()
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "alice@example.com") {
		t.Errorf("expected email in output, got: %q", out)
	}
	if !strings.Contains(out, "valid") {
		t.Errorf("expected 'valid' status in output, got: %q", out)
	}
}

func TestAuthStatus_ExpiredToken(t *testing.T) {
	isolateConfig(t)
	if err := auth.Save(&auth.TokenData{
		AccessToken: "tok",
		Email:       "alice@example.com",
		Expiry:      time.Now().Add(-time.Hour),
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var buf bytes.Buffer
	cmd := newAuthStatusCmd()
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(buf.String(), "EXPIRED") {
		t.Errorf("expected 'EXPIRED' in output, got: %q", buf.String())
	}
}

func TestAuthStatus_NotLoggedIn(t *testing.T) {
	isolateConfig(t)
	cmd := newAuthStatusCmd()
	cmd.SetOut(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when not logged in, got nil")
	}
}

func TestAuthStatus_EnvToken(t *testing.T) {
	isolateConfig(t)
	t.Setenv(EnvToken, "ci-bearer-token")

	var buf bytes.Buffer
	cmd := newAuthStatusCmd()
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status with env token: %v", err)
	}
	if !strings.Contains(buf.String(), "CI") {
		t.Errorf("expected CI mode in output, got: %q", buf.String())
	}
}
