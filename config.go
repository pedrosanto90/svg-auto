package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultConfigName = "config.json"

type FileRule struct {
	Name      string `json:"name"`
	Mode      string `json:"mode"`
	Marker    string `json:"marker"`
	Position  string `json:"position"`
	Template  string `json:"template"`
	Separator bool   `json:"separator"`
}

type Config struct {
	ProjectPath string     `json:"projectPath"`
	IconPrefix  string     `json:"iconPrefix"`
	Files       []FileRule `json:"files"`
}

func configPath() (string, error) {
	if p := os.Getenv("SVG_AUTO_CONFIG"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to find the config directory: %w", err)
	}
	return filepath.Join(dir, "svg-auto", defaultConfigName), nil
}

func loadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found at %s. Create it with: mkdir -p %s && nano %s", path, filepath.Dir(path), path)
		}
		return nil, fmt.Errorf("failed to read config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
	}
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", path, err)
	}
	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if strings.TrimSpace(cfg.ProjectPath) == "" {
		return fmt.Errorf("projectPath is required")
	}
	if cfg.IconPrefix == "" {
		cfg.IconPrefix = "icon-"
	}
	if len(cfg.Files) == 0 {
		return fmt.Errorf("files must list at least one file")
	}
	for i := range cfg.Files {
		f := &cfg.Files[i]
		if strings.TrimSpace(f.Name) == "" {
			return fmt.Errorf("each file needs a name")
		}
		switch f.Mode {
		case "text":
			if strings.TrimSpace(f.Template) == "" {
				return fmt.Errorf("file %q (text mode) needs a template", f.Name)
			}
			switch f.Position {
			case "before", "after", "replace", "end":
			default:
				return fmt.Errorf("file %q has an invalid position %q", f.Name, f.Position)
			}
			if f.Position != "end" && f.Marker == "" {
				return fmt.Errorf("file %q (position %q) needs a marker", f.Name, f.Position)
			}
		case "icomoon":
		default:
			return fmt.Errorf("file %q has an invalid mode %q (must be text or icomoon)", f.Name, f.Mode)
		}
	}
	return nil
}
