// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package main

import (
	"fmt"
	"os"

	"github.com/smeltry-io/smeltry/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
