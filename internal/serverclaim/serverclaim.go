// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package serverclaim provides read/write access to ServerClaim resources
// via the Kubernetes dynamic client.
package serverclaim

import (
	"context"
	"fmt"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructuredPkg "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var gvr = schema.GroupVersionResource{
	Group:    "portal.smeltry.io",
	Version:  "v1alpha1",
	Resource: "serverclaims",
}

// ServerClaim is a minimal representation of a ServerClaim resource.
type ServerClaim struct {
	Name      string
	Namespace string
	Phase     string
	Site      string
	Class     string
	OS        string
	ServerIP  string
	Age       string
}

// Client wraps a dynamic.Interface scoped to ServerClaim operations.
type Client struct {
	dyn dynamic.Interface
}

// NewClient returns a Client using the provided dynamic client.
func NewClient(dyn dynamic.Interface) *Client {
	return &Client{dyn: dyn}
}

// List returns all ServerClaims in namespace.
func (c *Client) List(ctx context.Context, namespace string) ([]ServerClaim, error) {
	list, err := c.dyn.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing ServerClaims: %w", err)
	}
	out := make([]ServerClaim, 0, len(list.Items))
	for _, item := range list.Items {
		out = append(out, fromUnstructured(item.Object))
	}
	return out, nil
}

// Get returns a single ServerClaim by name.
func (c *Client) Get(ctx context.Context, namespace, name string) (*ServerClaim, error) {
	obj, err := c.dyn.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting ServerClaim %q: %w", name, err)
	}
	sc := fromUnstructured(obj.Object)
	return &sc, nil
}

// Delete deletes a ServerClaim by name.
func (c *Client) Delete(ctx context.Context, namespace, name string) error {
	err := c.dyn.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("deleting ServerClaim %q: %w", name, err)
	}
	return nil
}

// IsGone returns true when the ServerClaim no longer exists.
func (c *Client) IsGone(ctx context.Context, namespace, name string) (bool, error) {
	_, err := c.dyn.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return false, nil
	}
	if k8serrors.IsNotFound(err) {
		return true, nil
	}
	return false, err
}

// Create applies an unstructured ServerClaim object.
func (c *Client) Create(ctx context.Context, namespace string, obj map[string]interface{}) (*ServerClaim, error) {
	result, err := c.dyn.Resource(gvr).Namespace(namespace).Create(ctx,
		&unstructuredPkg.Unstructured{Object: obj}, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("creating ServerClaim: %w", err)
	}
	sc := fromUnstructured(result.Object)
	return &sc, nil
}

func fromUnstructured(obj map[string]interface{}) ServerClaim {
	sc := ServerClaim{
		Name:      strField(obj, "metadata", "name"),
		Namespace: strField(obj, "metadata", "namespace"),
		Phase:     strField(obj, "status", "phase"),
		Site:      strField(obj, "spec", "site"),
		Class:     strField(obj, "spec", "machineClass"),
		OS:        strField(obj, "spec", "os"),
		ServerIP:  strField(obj, "status", "serverIP"),
	}
	if ts := strField(obj, "metadata", "creationTimestamp"); ts != "" {
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			sc.Age = humanAge(t)
		}
	}
	return sc
}

func humanAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func strField(obj map[string]interface{}, keys ...string) string {
	cur := obj
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return ""
		}
		if i == len(keys)-1 {
			s, _ := v.(string)
			return s
		}
		cur, _ = v.(map[string]interface{})
		if cur == nil {
			return ""
		}
	}
	return ""
}
