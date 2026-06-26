// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

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
