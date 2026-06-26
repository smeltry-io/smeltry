// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package poller provides a generic polling helper for --wait flags.
package poller

import (
	"context"
	"fmt"
	"time"
)

const DefaultInterval = 2 * time.Second

// UntilDone polls cond every interval until it returns (true, nil), the
// context is cancelled, or the context deadline is exceeded.
// It returns the last non-nil error from cond if the context expires.
func UntilDone(ctx context.Context, interval time.Duration, cond func(ctx context.Context) (done bool, err error)) error {
	var lastErr error
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return fmt.Errorf("timed out waiting: %w", lastErr)
			}
			return fmt.Errorf("timed out waiting: %w", ctx.Err())
		case <-ticker.C:
			done, err := cond(ctx)
			if err != nil {
				lastErr = err
				continue
			}
			if done {
				return nil
			}
		}
	}
}
