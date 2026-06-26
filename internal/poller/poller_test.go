// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package poller

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestUntilDone_SucceedsImmediately(t *testing.T) {
	calls := 0
	start := time.Now()
	err := UntilDone(context.Background(), time.Hour, func(_ context.Context) (bool, error) {
		calls++
		return true, nil
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
	// With interval=1h and immediate success, the function must return in well under 1s.
	if elapsed > time.Second {
		t.Errorf("expected immediate return, took %v", elapsed)
	}
}

func TestUntilDone_SucceedsAfterRetries(t *testing.T) {
	calls := 0
	err := UntilDone(context.Background(), 10*time.Millisecond, func(_ context.Context) (bool, error) {
		calls++
		return calls >= 3, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls < 3 {
		t.Errorf("expected at least 3 calls, got %d", calls)
	}
}

func TestUntilDone_TimesOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	sentinel := errors.New("still waiting")
	err := UntilDone(ctx, 10*time.Millisecond, func(_ context.Context) (bool, error) {
		return false, sentinel
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected last error to be wrapped in timeout error, got: %v", err)
	}
}

func TestUntilDone_ContextCancelledWithNoError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := UntilDone(ctx, 10*time.Millisecond, func(_ context.Context) (bool, error) {
		return false, nil
	})
	if err == nil {
		t.Fatal("expected error on timeout, got nil")
	}
}
