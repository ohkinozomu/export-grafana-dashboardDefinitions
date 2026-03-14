package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseDefinitions(t *testing.T) {
	body := []byte(`
data:
  sample.json: |
    {"title":"demo","panels":[]}
`)

	defs, err := parseDefinitions(body)
	if err != nil {
		t.Fatalf("parseDefinitions returned error: %v", err)
	}

	got := defs.Data["sample.json"]
	want := "{\"title\":\"demo\",\"panels\":[]}\n"
	if got != want {
		t.Fatalf("unexpected dashboard body: got %q want %q", got, want)
	}
}

func TestParseDefinitionsFromItems(t *testing.T) {
	body := []byte(`
apiVersion: v1
kind: List
items:
  - apiVersion: v1
    kind: ConfigMap
    data:
      sample.json: |
        {"title":"demo","panels":[]}
`)

	defs, err := parseDefinitions(body)
	if err != nil {
		t.Fatalf("parseDefinitions returned error: %v", err)
	}

	got := defs.Data["sample.json"]
	want := "{\"title\":\"demo\",\"panels\":[]}\n"
	if got != want {
		t.Fatalf("unexpected dashboard body: got %q want %q", got, want)
	}
}

func TestWriteDashboard(t *testing.T) {
	dir := t.TempDir()

	if err := writeDashboard(dir, "sample.json", `{"title":"demo","panels":[]}`); err != nil {
		t.Fatalf("writeDashboard returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "sample.json"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	var got any
	if err := json.Unmarshal(content, &got); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	want := map[string]any{"title": "demo", "panels": []any{}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected dashboard content: got %#v want %#v", got, want)
	}
}
