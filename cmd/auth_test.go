// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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

	if _, err := auth.Load(); err == nil {
		t.Error("expected error after logout, got nil")
	}
}

func TestAuthLogout_Idempotent(t *testing.T) {
	isolateConfig(t)
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

// ── extractIDTokenClaims ────────────────────────────────────────────────────

func makeIDToken(t *testing.T, claims map[string]interface{}) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload, _ := json.Marshal(claims)
	body := base64.RawURLEncoding.EncodeToString(payload)
	return strings.Join([]string{header, body, "sig"}, ".")
}

func TestExtractIDTokenClaims_AllFields(t *testing.T) {
	exp := time.Now().Add(time.Hour).Unix()
	tok := makeIDToken(t, map[string]interface{}{
		"email":  "user@example.com",
		"groups": []string{"tenant-acme", "smeltry-admins"},
		"exp":    exp,
	})

	email, groups, expiry, err := extractIDTokenClaims(tok, 3600)
	if err != nil {
		t.Fatalf("extractIDTokenClaims: %v", err)
	}
	if email != "user@example.com" {
		t.Errorf("email: got %q want %q", email, "user@example.com")
	}
	if len(groups) != 2 {
		t.Errorf("groups: got %v want 2 items", groups)
	}
	if expiry.Unix() != exp {
		t.Errorf("expiry: got %v want %v", expiry.Unix(), exp)
	}
}

func TestExtractIDTokenClaims_FallbackExpiry(t *testing.T) {
	tok := makeIDToken(t, map[string]interface{}{
		"email": "user@example.com",
	})
	before := time.Now()
	_, _, expiry, err := extractIDTokenClaims(tok, 3600)
	if err != nil {
		t.Fatalf("extractIDTokenClaims: %v", err)
	}
	expected := before.Add(3600 * time.Second)
	if expiry.Before(before) || expiry.After(expected.Add(5*time.Second)) {
		t.Errorf("fallback expiry out of expected range: %v", expiry)
	}
}

func TestExtractIDTokenClaims_MalformedToken(t *testing.T) {
	_, _, _, err := extractIDTokenClaims("not.a.valid.jwt.here", 3600)
	if err == nil {
		t.Error("expected error for malformed token (4 parts)")
	}
}

func TestExtractIDTokenClaims_BadBase64(t *testing.T) {
	_, _, _, err := extractIDTokenClaims("hdr.!!!.sig", 3600)
	if err == nil {
		t.Error("expected error for bad base64 payload")
	}
}
