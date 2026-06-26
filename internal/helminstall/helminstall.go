// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The Smeltry Authors

// Package helminstall installs or upgrades the smeltry-operator Helm chart
// using the Helm Go SDK (no helm binary required).
package helminstall

import (
	"fmt"
	"io"
	"log"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
)

const (
	// DefaultChart is the OCI reference for the smeltry-operator chart.
	DefaultChart = "oci://ghcr.io/smeltry-io/helm-charts/smeltry-operator"
	// DefaultRelease is the Helm release name used by install/upgrade.
	DefaultRelease = "smeltry-operator"
)

// Options configures an install or upgrade operation.
type Options struct {
	Namespace   string
	Version     string
	ValueFiles  []string
	SetValues   []string
	DryRun      bool
	Wait        bool
	Timeout     time.Duration
	KubeContext string
}

// Install installs or upgrades the smeltry-operator chart on the cluster.
// Output (Helm notes, status) is written to w.
func Install(w io.Writer, opts Options) error {
	env := cli.New()
	env.SetNamespace(opts.Namespace)
	if opts.KubeContext != "" {
		env.KubeContext = opts.KubeContext
	}

	rc, err := registry.NewClient(registry.ClientOptWriter(io.Discard))
	if err != nil {
		return fmt.Errorf("creating registry client: %w", err)
	}

	cfg := new(action.Configuration)
	if err := cfg.Init(env.RESTClientGetter(), opts.Namespace, "secret",
		func(format string, v ...interface{}) {
			log.Printf(format, v...)
		}); err != nil {
		return fmt.Errorf("initialising Helm configuration: %w", err)
	}
	cfg.RegistryClient = rc

	vals, err := buildValues(opts.ValueFiles, opts.SetValues)
	if err != nil {
		return err
	}

	ch, err := loadChart(cfg, opts.Version)
	if err != nil {
		return err
	}

	// Check whether the release already exists to decide install vs upgrade.
	hist := action.NewHistory(cfg)
	hist.Max = 1
	_, err = hist.Run(DefaultRelease)
	exists := err == nil

	var rel *release.Release
	if exists {
		rel, err = runUpgrade(cfg, ch, opts, vals)
	} else {
		rel, err = runInstall(cfg, ch, opts, vals)
	}
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Release %q deployed to namespace %q (status: %s)\n",
		rel.Name, rel.Namespace, rel.Info.Status)
	if rel.Info.Notes != "" {
		fmt.Fprintln(w, rel.Info.Notes)
	}
	return nil
}

func runInstall(cfg *action.Configuration, ch *chart.Chart, opts Options, vals map[string]interface{}) (*release.Release, error) {
	act := action.NewInstall(cfg)
	act.ReleaseName = DefaultRelease
	act.Namespace = opts.Namespace
	act.CreateNamespace = true
	act.DryRun = opts.DryRun
	act.Wait = opts.Wait
	act.Timeout = opts.Timeout
	act.Version = opts.Version
	return act.Run(ch, vals)
}

func runUpgrade(cfg *action.Configuration, ch *chart.Chart, opts Options, vals map[string]interface{}) (*release.Release, error) {
	act := action.NewUpgrade(cfg)
	act.Namespace = opts.Namespace
	act.DryRun = opts.DryRun
	act.Wait = opts.Wait
	act.Timeout = opts.Timeout
	act.Version = opts.Version
	act.ReuseValues = false
	return act.Run(DefaultRelease, ch, vals)
}

// loadChart pulls the smeltry-operator chart from the OCI registry.
func loadChart(cfg *action.Configuration, version string) (*chart.Chart, error) {
	act := action.NewInstall(cfg)
	act.Version = version
	cp, err := act.ChartPathOptions.LocateChart(DefaultChart, cli.New())
	if err != nil {
		return nil, fmt.Errorf("locating chart %q: %w", DefaultChart, err)
	}
	ch, err := loader.Load(cp)
	if err != nil {
		return nil, fmt.Errorf("loading chart: %w", err)
	}
	return ch, nil
}

// buildValues merges values files and --set overrides (set wins over files).
func buildValues(valueFiles, setVals []string) (map[string]interface{}, error) {
	opts := &values.Options{
		ValueFiles: valueFiles,
		Values:     setVals,
	}
	providers := getter.All(cli.New())
	return opts.MergeValues(providers)
}

// chartRef returns the OCI chart reference. The version is passed separately
// to Helm actions rather than embedded in the ref.
func chartRef(_ string) string {
	return DefaultChart
}
