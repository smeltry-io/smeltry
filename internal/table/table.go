// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package table wraps tablewriter with the project's default rendering style.
package table

import (
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// New returns a tablewriter preconfigured with the smeltry style:
// no visible borders, space-separated columns, clean header.
func New(w io.Writer) *tablewriter.Table {
	t := tablewriter.NewWriter(w)
	t.Configure(func(cfg *tablewriter.Config) {
		cfg.Behavior.TrimSpace = tw.Success
		cfg.Header.Alignment.Global = tw.AlignLeft
		cfg.Row.Alignment.Global = tw.AlignLeft
	})
	t.Options(tablewriter.WithBorders(tw.Border{
		Left:   tw.Fail,
		Right:  tw.Fail,
		Top:    tw.Fail,
		Bottom: tw.Fail,
	}))
	return t
}

// NewStdout is a convenience wrapper writing to os.Stdout.
func NewStdout() *tablewriter.Table {
	return New(os.Stdout)
}
