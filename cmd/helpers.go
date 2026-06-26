// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// notImplemented is a placeholder RunE for commands not yet implemented.
func notImplemented(cmd *cobra.Command, _ []string) error {
	return fmt.Errorf("%q is not yet implemented", cmd.CommandPath())
}
