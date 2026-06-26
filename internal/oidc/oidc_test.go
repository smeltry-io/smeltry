// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// fakeOIDCServer returns a test server that simulates Authentik endpoints.
func fakeOIDCServer(t *testing.T, pollResponses []map[string]interface{}) *httptest.Server {
	t.Helper()
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(map[string]string{
				"device_authorization_endpoint": "http://" + r.Host + "/device",
				"token_endpoint":                "http://" + r.Host + "/token",
			})
		case "/device":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"device_code":      "dev-code-123",
				"user_code":        "ABCD-1234",
				"verification_uri": "https://auth.example.com/activate",
				"expires_in":       300,
				"interval":         1,
			})
		case "/token":
			idx := int(callCount.Load())
			resp := pollResponses[idx]
			if idx < len(pollResponses)-1 {
				callCount.Add(1)
			}
			if errVal, ok := resp["error"]; ok {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": errVal.(string)})
			} else {
				json.NewEncoder(w).Encode(resp)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestDiscover_ParsesEndpoints(t *testing.T) {
	srv := fakeOIDCServer(t, nil)
	c := New()
	doc, err := c.Discover(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if doc.DeviceAuthEndpoint == "" {
		t.Error("expected device_authorization_endpoint to be set")
	}
	if doc.TokenEndpoint == "" {
		t.Error("expected token_endpoint to be set")
	}
}

func TestDiscover_ErrorOnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New()
	_, err := c.Discover(context.Background(), srv.URL)
	if err == nil {
		t.Error("expected error on non-200 status")
	}
}

func TestStartDeviceAuth_ReturnsUserCode(t *testing.T) {
	srv := fakeOIDCServer(t, nil)
	c := New()
	doc, _ := c.Discover(context.Background(), srv.URL)
	dar, err := c.StartDeviceAuth(context.Background(), doc.DeviceAuthEndpoint, "smeltry-cli", []string{"openid", "email"})
	if err != nil {
		t.Fatalf("StartDeviceAuth: %v", err)
	}
	if dar.UserCode != "ABCD-1234" {
		t.Errorf("UserCode: got %q want %q", dar.UserCode, "ABCD-1234")
	}
	if dar.DeviceCode != "dev-code-123" {
		t.Errorf("DeviceCode: got %q want %q", dar.DeviceCode, "dev-code-123")
	}
	if dar.VerificationURI == "" {
		t.Error("VerificationURI should be set")
	}
	if dar.Interval == 0 {
		t.Error("Interval should default to non-zero")
	}
}

func TestPollToken_SuccessAfterPending(t *testing.T) {
	responses := []map[string]interface{}{
		{"error": "authorization_pending"},
		{"error": "authorization_pending"},
		{
			"access_token":  "access-tok",
			"refresh_token": "refresh-tok",
			"expires_in":    3600,
			"id_token":      "id-tok",
		},
	}
	srv := fakeOIDCServer(t, responses)
	c := New()
	doc, _ := c.Discover(context.Background(), srv.URL)

	tr, err := c.PollToken(context.Background(), doc.TokenEndpoint, "smeltry-cli", "dev-code-123", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("PollToken: %v", err)
	}
	if tr.AccessToken != "access-tok" {
		t.Errorf("AccessToken: got %q want %q", tr.AccessToken, "access-tok")
	}
	if tr.RefreshToken != "refresh-tok" {
		t.Errorf("RefreshToken: got %q want %q", tr.RefreshToken, "refresh-tok")
	}
}

func TestPollToken_SlowDownIncreasesInterval(t *testing.T) {
	responses := []map[string]interface{}{
		{"error": "slow_down"},
		{"access_token": "tok", "refresh_token": "rtok", "expires_in": 3600, "id_token": "id"},
	}
	srv := fakeOIDCServer(t, responses)
	c := New()
	doc, _ := c.Discover(context.Background(), srv.URL)

	// Should succeed despite slow_down (interval increases internally).
	tr, err := c.PollToken(context.Background(), doc.TokenEndpoint, "smeltry-cli", "dev-code-123", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("PollToken slow_down: %v", err)
	}
	if tr.AccessToken != "tok" {
		t.Errorf("AccessToken: got %q want %q", tr.AccessToken, "tok")
	}
}

func TestPollToken_AccessDenied(t *testing.T) {
	responses := []map[string]interface{}{{"error": "access_denied"}}
	srv := fakeOIDCServer(t, responses)
	c := New()
	doc, _ := c.Discover(context.Background(), srv.URL)

	_, err := c.PollToken(context.Background(), doc.TokenEndpoint, "smeltry-cli", "dev-code-123", 10*time.Millisecond)
	if err == nil {
		t.Error("expected error on access_denied")
	}
}

func TestPollToken_ExpiredToken(t *testing.T) {
	responses := []map[string]interface{}{{"error": "expired_token"}}
	srv := fakeOIDCServer(t, responses)
	c := New()
	doc, _ := c.Discover(context.Background(), srv.URL)

	_, err := c.PollToken(context.Background(), doc.TokenEndpoint, "smeltry-cli", "dev-code-123", 10*time.Millisecond)
	if err == nil {
		t.Error("expected error on expired_token")
	}
}

func TestPollToken_ContextCancelled(t *testing.T) {
	responses := []map[string]interface{}{{"error": "authorization_pending"}}
	srv := fakeOIDCServer(t, responses)
	c := New()
	doc, _ := c.Discover(context.Background(), srv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := c.PollToken(ctx, doc.TokenEndpoint, "smeltry-cli", "dev-code-123", 10*time.Millisecond)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}
