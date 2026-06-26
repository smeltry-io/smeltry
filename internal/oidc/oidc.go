// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package oidc implements the OAuth2 Device Authorization Grant (RFC 8628).
package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client performs OIDC operations over HTTP.
type Client struct {
	HTTP *http.Client
}

// New returns a Client using the default HTTP client.
func New() *Client {
	return &Client{HTTP: http.DefaultClient}
}

// DiscoveryDoc holds the subset of the OIDC discovery document needed for
// the device flow.
type DiscoveryDoc struct {
	DeviceAuthEndpoint string `json:"device_authorization_endpoint"`
	TokenEndpoint      string `json:"token_endpoint"`
}

// Discover fetches the OIDC discovery document from issuerURL.
func (c *Client) Discover(ctx context.Context, issuerURL string) (*DiscoveryDoc, error) {
	u := strings.TrimRight(issuerURL, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("building discovery request: %w", err)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching discovery document: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery returned HTTP %d", resp.StatusCode)
	}
	var doc DiscoveryDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("parsing discovery document: %w", err)
	}
	return &doc, nil
}

// DeviceAuthResponse is the response from the device authorization endpoint.
type DeviceAuthResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// StartDeviceAuth initiates the device authorization flow.
func (c *Client) StartDeviceAuth(ctx context.Context, endpoint, clientID string, scopes []string) (*DeviceAuthResponse, error) {
	vals := url.Values{
		"client_id": {clientID},
		"scope":     {strings.Join(scopes, " ")},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, fmt.Errorf("building device auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device authorization request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device authorization returned HTTP %d", resp.StatusCode)
	}
	var dar DeviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&dar); err != nil {
		return nil, fmt.Errorf("parsing device auth response: %w", err)
	}
	if dar.Interval == 0 {
		dar.Interval = 5 // RFC 8628 §3.5 default
	}
	return &dar, nil
}

// TokenResponse holds the tokens returned after successful device authorization.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
}

type tokenErrorResponse struct {
	Error string `json:"error"`
}

// PollToken polls the token endpoint until the user completes the device flow,
// the context is cancelled, or the device code expires.
// interval is the initial polling interval; it grows on slow_down responses.
func (c *Client) PollToken(ctx context.Context, tokenEndpoint, clientID, deviceCode string, interval time.Duration) (*TokenResponse, error) {
	vals := url.Values{
		"client_id":   {clientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(vals.Encode()))
		if err != nil {
			return nil, fmt.Errorf("building token request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			return nil, fmt.Errorf("token request: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			var tr TokenResponse
			err := json.NewDecoder(resp.Body).Decode(&tr)
			resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("parsing token response: %w", err)
			}
			return &tr, nil
		}

		var errResp tokenErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp) //nolint:errcheck
		resp.Body.Close()

		switch errResp.Error {
		case "authorization_pending":
			// keep polling at current interval
		case "slow_down":
			// RFC 8628 §3.5: increase interval by 5 seconds
			interval += 5 * time.Second
		case "access_denied":
			return nil, fmt.Errorf("access denied by user")
		case "expired_token":
			return nil, fmt.Errorf("device code expired — run 'smeltry auth login' again")
		default:
			return nil, fmt.Errorf("token endpoint error: %q", errResp.Error)
		}
	}
}
