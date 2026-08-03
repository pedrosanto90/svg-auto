package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	t.Setenv("SVG_AUTO_CONFIG", path)
	return path
}

func TestLoadConfig(t *testing.T) {
	writeConfig(t, `{
  "projectPath": "user@server:/srv/www/app",
  "iconPrefix": "ee-icon-",
  "files": [
    { "name": "a.svg", "mode": "text", "marker": "</defs>", "position": "before", "template": "<symbol id=\"{{.Prefix}}{{.Name}}\" viewBox=\"{{.ViewBox}}\">{{.Body}}</symbol>" },
    { "name": "a.json", "mode": "icomoon" },
    { "name": "a.css", "mode": "text", "position": "end", "template": ".{{.Prefix}}{{.Name}} { width: 1.75em; }" }
  ]
}`)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ProjectPath != "user@server:/srv/www/app" {
		t.Errorf("unexpected projectPath: %s", cfg.ProjectPath)
	}
	if cfg.IconPrefix != "ee-icon-" {
		t.Errorf("unexpected iconPrefix: %s", cfg.IconPrefix)
	}
	if len(cfg.Files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(cfg.Files))
	}
	if cfg.Files[0].Position != "before" || cfg.Files[0].Marker != "</defs>" {
		t.Errorf("unexpected first rule: %+v", cfg.Files[0])
	}
}

func TestLoadConfigMissing(t *testing.T) {
	t.Setenv("SVG_AUTO_CONFIG", filepath.Join(t.TempDir(), "nope.json"))
	if _, err := loadConfig(); err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestValidateConfigDefaultPrefix(t *testing.T) {
	cfg := &Config{
		ProjectPath: "/srv/app",
		Files: []FileRule{
			{Name: "a.svg", Mode: "text", Marker: "x", Position: "before", Template: "{{.Name}}"},
		},
	}
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IconPrefix != "icon-" {
		t.Errorf("expected default prefix icon-, got %q", cfg.IconPrefix)
	}
}

func TestValidateConfigErrors(t *testing.T) {
	cases := []Config{
		{ProjectPath: "", Files: []FileRule{{Name: "a", Mode: "text", Marker: "x", Position: "before", Template: "t"}}},
		{ProjectPath: "/x", Files: nil},
		{ProjectPath: "/x", Files: []FileRule{{Name: "a", Mode: "bogus"}}},
		{ProjectPath: "/x", Files: []FileRule{{Name: "a", Mode: "text", Position: "before", Template: "t"}}},
		{ProjectPath: "/x", Files: []FileRule{{Name: "a", Mode: "text", Marker: "x", Position: "sideways", Template: "t"}}},
		{ProjectPath: "/x", Files: []FileRule{{Name: "a", Mode: "text", Marker: "x", Position: "before"}}},
	}
	for i, c := range cases {
		if err := validateConfig(&c); err == nil {
			t.Errorf("case %d: expected error, got nil", i)
		}
	}
}
