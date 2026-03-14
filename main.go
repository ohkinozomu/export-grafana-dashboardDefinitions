package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

const defaultBaseURL = "https://raw.githubusercontent.com/prometheus-operator/kube-prometheus/%s/manifests/grafana-dashboardDefinitions.yaml"

type dashboardDefinitions struct {
	Data  map[string]string `yaml:"data"`
	Items []struct {
		Data map[string]string `yaml:"data"`
	} `yaml:"items"`
}

func main() {
	var (
		ref     = flag.String("ref", "refs/heads/main", "Git ref to fetch, for example refs/heads/main or refs/tags/v0.14.0")
		outDir  = flag.String("out", "out", "Directory where dashboard JSON files will be written")
		baseURL = flag.String("base-url", defaultBaseURL, "Printf-style URL template with a single %s placeholder for ref")
	)
	flag.Parse()

	if err := run(*ref, *outDir, *baseURL); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ref, outDir, baseURL string) error {
	if !strings.Contains(baseURL, "%s") {
		return errors.New("base-url must contain a %s placeholder for ref")
	}

	body, err := fetchManifest(fmt.Sprintf(baseURL, ref))
	if err != nil {
		return err
	}

	defs, err := parseDefinitions(body)
	if err != nil {
		return err
	}

	if len(defs.Data) == 0 {
		return errors.New("manifest data is empty")
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	for name, raw := range defs.Data {
		if err := writeDashboard(outDir, name, raw); err != nil {
			return err
		}
	}

	fmt.Printf("exported %d dashboards to %s\n", len(defs.Data), outDir)
	return nil
}

func fetchManifest(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch manifest: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	return body, nil
}

func parseDefinitions(body []byte) (*dashboardDefinitions, error) {
	var defs dashboardDefinitions
	if err := yaml.Unmarshal(body, &defs); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if len(defs.Data) == 0 && len(defs.Items) > 0 {
		merged := make(map[string]string)
		for _, item := range defs.Items {
			for name, raw := range item.Data {
				merged[name] = raw
			}
		}
		defs.Data = merged
	}

	return &defs, nil
}

func writeDashboard(outDir, name, raw string) error {
	var pretty any
	if err := json.Unmarshal([]byte(raw), &pretty); err != nil {
		return fmt.Errorf("parse dashboard %q: %w", name, err)
	}

	content, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		return fmt.Errorf("format dashboard %q: %w", name, err)
	}
	content = append(content, '\n')

	target := filepath.Join(outDir, filepath.Base(name))
	if err := os.WriteFile(target, content, 0o644); err != nil {
		return fmt.Errorf("write dashboard %q: %w", name, err)
	}
	return nil
}
