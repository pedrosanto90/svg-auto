package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testSVG = `<svg aria-hidden="true" style="position: absolute; width: 0; height: 0; overflow: hidden;" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
<defs>
<symbol id="icon-house-solid" viewBox="0 0 32 32">
<path d="M17.363 0.537c-0.769-0.713-1.956-0.713-2.719 0l-14 13c-0.6 0.563-0.8 1.431-0.5 2.194s1.031 1.269 1.856 1.269h1v11c0 2.206 1.794 4 4 4h18c2.206 0 4-1.794 4-4v-11h1c0.825 0 1.563-0.506 1.863-1.269s0.1-1.637-0.5-2.194l-14-13zM15 20h2c1.656 0 3 1.344 3 3v6h-8v-6c0-1.656 1.344-3 3-3z"></path>
</symbol>
<symbol id="icon-search" viewBox="0 0 24 24">
<path d="M10 0"></path>
<path d="M20 0"></path>
</symbol>
</defs>
</svg>
`

const testJSON = `{
  "IcoMoonType": "selection",
  "icons": [
    {
      "icon": {
        "paths": [
          "M555.6 17.2c-24.6-22.8-62.6-22.8-87 0l-448 416c-19.2 18-25.6 45.8-16 70.2s33 40.6 59.4 40.6h32v352c0 70.6 57.4 128 128 128h576c70.6 0 128-57.4 128-128v-352h32c26.4 0 50-16.2 59.6-40.6s3.2-52.4-16-70.2l-448-416zM480 640h64c53 0 96 43 96 96v192h-256v-192c0-53 43-96 96-96z"
        ]
      }
    },
    {
      "icon": {
        "paths": [
          "M10 0c1 2",
          "M20 0c3 4"
        ]
      }
    }
  ]
}
`

func TestExtractSymbols(t *testing.T) {
	icons, err := extractSymbols([]byte(testSVG))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(icons) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(icons))
	}

	first := icons[0]
	if first.Name != "house-solid" {
		t.Errorf("unexpected name: %s", first.Name)
	}
	if first.ViewBox != "0 0 32 32" {
		t.Errorf("unexpected viewBox: %s", first.ViewBox)
	}
	if !strings.Contains(first.Body, `<path d=`) || !strings.HasSuffix(first.Body, `</path>`) {
		t.Errorf("unexpected body: %s", first.Body)
	}

	second := icons[1]
	if second.Name != "search" {
		t.Errorf("unexpected name: %s", second.Name)
	}
	if second.ViewBox != "0 0 24 24" {
		t.Errorf("unexpected viewBox: %s", second.ViewBox)
	}
	if strings.Count(second.Body, `<path `) != 2 {
		t.Errorf("expected 2 paths in body, got: %s", second.Body)
	}
}

func TestExtractSymbolsEmpty(t *testing.T) {
	if _, err := extractSymbols([]byte(`<svg></svg>`)); err == nil {
		t.Fatal("expected error for empty svg, got nil")
	}
}

func TestExtractPaths(t *testing.T) {
	paths, err := extractPaths([]byte(testJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 icon entries, got %d", len(paths))
	}
	if len(paths[0]) != 1 || !strings.HasPrefix(paths[0][0], "M555.6 17.2c") {
		t.Errorf("unexpected first paths: %v", paths[0])
	}
	if len(paths[1]) != 2 {
		t.Errorf("expected 2 paths in second entry, got %v", paths[1])
	}
}

func TestExtractPathsEmpty(t *testing.T) {
	if _, err := extractPaths([]byte(`{"icons": [{"icon": {"paths": []}}]}`)); err == nil {
		t.Fatal("expected error for empty paths, got nil")
	}
	if _, err := extractPaths([]byte(`not json`)); err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}

func writeTestZip(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("failed to create entry %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write entry %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}
	return path
}

func TestProcessZip(t *testing.T) {
	zipPath := writeTestZip(t, map[string]string{
		"symbol-defs.svg": testSVG,
		"selection.json":  testJSON,
	})

	icons, err := processZip(zipPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(icons) != 2 {
		t.Fatalf("expected 2 icons, got %d", len(icons))
	}
	if icons[0].Name != "house-solid" || len(icons[0].Paths) != 1 {
		t.Errorf("unexpected first icon: %+v", icons[0])
	}
	if icons[1].Name != "search" || len(icons[1].Paths) != 2 {
		t.Errorf("unexpected second icon: %+v", icons[1])
	}
}

func TestReadZipEntryMissing(t *testing.T) {
	zipPath := writeTestZip(t, map[string]string{
		"selection.json": testJSON,
	})
	if _, err := readZipEntry(zipPath, "symbol-defs.svg"); err == nil {
		t.Fatal("expected error for missing entry, got nil")
	}
}
