// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// notImplemented is a placeholder RunE for commands not yet implemented.
func notImplemented(cmd *cobra.Command, _ []string) error {
	return fmt.Errorf("%q is not yet implemented", cmd.CommandPath())
}

// loadManifest reads a YAML or JSON manifest file and returns the parsed object
// and the namespace declared in metadata (if any).
func loadManifest(path string) (obj map[string]interface{}, ns string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("reading file %q: %w", path, err)
	}
	jsonBytes, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, "", fmt.Errorf("parsing manifest: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		return nil, "", fmt.Errorf("unmarshalling manifest: %w", err)
	}
	if meta, ok := obj["metadata"].(map[string]interface{}); ok {
		ns, _ = meta["namespace"].(string)
	}
	return obj, ns, nil
}
