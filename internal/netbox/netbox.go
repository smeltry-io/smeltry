// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package netbox provides a minimal HTTP client for the Netbox REST API.
package netbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Client is a minimal Netbox API client.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewClient returns a Client for the given Netbox base URL and API token.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    http.DefaultClient,
	}
}

// DeviceStatus holds the status value of a Netbox device.
type DeviceStatus struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// NestedSite holds the site slug of a Netbox device.
type NestedSite struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// NestedType holds the device type model name.
type NestedType struct {
	Model string `json:"model"`
}

// Device is a Netbox dcim device.
type Device struct {
	ID         int          `json:"id"`
	Name       string       `json:"name"`
	Status     DeviceStatus `json:"status"`
	Site       NestedSite   `json:"site"`
	DeviceType NestedType   `json:"device_type"`
}

// ListDevicesParams are optional filters for ListDevices.
type ListDevicesParams struct {
	Site   string
	Status string
}

type listResponse struct {
	Count   int      `json:"count"`
	Results []Device `json:"results"`
}

// ListDevices returns all devices matching the given filters.
func (c *Client) ListDevices(ctx context.Context, params ListDevicesParams) ([]Device, error) {
	u, err := url.Parse(c.baseURL + "/api/dcim/devices/")
	if err != nil {
		return nil, fmt.Errorf("building URL: %w", err)
	}
	q := u.Query()
	if params.Site != "" {
		q.Set("site", params.Site)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Token "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("netbox request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("netbox returned HTTP %d", resp.StatusCode)
	}

	var result listResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return result.Results, nil
}
