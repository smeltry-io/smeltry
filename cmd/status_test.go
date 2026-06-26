// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/smeltry-io/smeltry/internal/operator"
)

func TestFormatStatus_OperatorReady(t *testing.T) {
	st := operator.Status{
		Ready:        true,
		Replicas:     1,
		ReadyReplicas: 1,
		Version:      "v0.1.0",
		ClusterClaims: 3,
		ServerClaims:  2,
	}
	var buf bytes.Buffer
	formatStatus(&buf, st)
	out := buf.String()

	if !strings.Contains(out, "Ready") {
		t.Errorf("expected 'Ready' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "v0.1.0") {
		t.Errorf("expected version in output, got:\n%s", out)
	}
	if !strings.Contains(out, "ClusterClaims:  3") {
		t.Errorf("expected ClusterClaims count in output, got:\n%s", out)
	}
	if !strings.Contains(out, "ServerClaims:   2") {
		t.Errorf("expected ServerClaims count in output, got:\n%s", out)
	}
}

func TestFormatStatus_OperatorNotReady(t *testing.T) {
	st := operator.Status{
		Ready:         false,
		Replicas:      1,
		ReadyReplicas: 0,
		Version:       "",
	}
	var buf bytes.Buffer
	formatStatus(&buf, st)
	out := buf.String()

	if !strings.Contains(out, "Not Ready") {
		t.Errorf("expected 'Not Ready' in output, got:\n%s", out)
	}
}

func TestFormatStatus_OperatorAbsent(t *testing.T) {
	st := operator.Status{
		Ready:    false,
		Absent:   true,
	}
	var buf bytes.Buffer
	formatStatus(&buf, st)
	out := buf.String()

	if !strings.Contains(out, "not installed") {
		t.Errorf("expected 'not installed' in output, got:\n%s", out)
	}
}
