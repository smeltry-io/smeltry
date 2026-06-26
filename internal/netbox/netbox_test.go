// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package netbox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeNetbox(t *testing.T, devices []Device) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dcim/devices/" {
			http.NotFound(w, r)
			return
		}
		resp := map[string]interface{}{
			"count":   len(devices),
			"results": devices,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestListDevices_ReturnsMachines(t *testing.T) {
	devices := []Device{
		{ID: 1, Name: "server-01", Status: DeviceStatus{Value: "active"}, Site: NestedSite{Slug: "paris-dc1"}, DeviceType: NestedType{Model: "Dell R740"}},
		{ID: 2, Name: "server-02", Status: DeviceStatus{Value: "staged"}, Site: NestedSite{Slug: "paris-dc1"}, DeviceType: NestedType{Model: "Dell R740"}},
	}
	srv := fakeNetbox(t, devices)

	c := NewClient(srv.URL, "test-token")
	got, err := c.ListDevices(context.Background(), ListDevicesParams{})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(got))
	}
	if got[0].Name != "server-01" {
		t.Errorf("Name[0]: got %q want %q", got[0].Name, "server-01")
	}
	if got[1].Status.Value != "staged" {
		t.Errorf("Status[1]: got %q want %q", got[1].Status.Value, "staged")
	}
}

func TestListDevices_EmptyResult(t *testing.T) {
	srv := fakeNetbox(t, []Device{})
	c := NewClient(srv.URL, "test-token")
	got, err := c.ListDevices(context.Background(), ListDevicesParams{})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 devices, got %d", len(got))
	}
}

func TestListDevices_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "bad-token")
	_, err := c.ListDevices(context.Background(), ListDevicesParams{})
	if err == nil {
		t.Error("expected error on 401")
	}
}

func TestListDevices_SiteFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		site := r.URL.Query().Get("site")
		if site != "paris-dc1" {
			http.Error(w, "unexpected site filter", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"count": 0, "results": []Device{}})
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "tok")
	_, err := c.ListDevices(context.Background(), ListDevicesParams{Site: "paris-dc1"})
	if err != nil {
		t.Fatalf("ListDevices with site filter: %v", err)
	}
}
