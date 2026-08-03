package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testIcons() []Icon {
	return []Icon{
		{Name: "house-solid", ViewBox: "0 0 32 32", Body: `<path d="M17.363 0.537"></path>`, Paths: []string{"M555.6 17.2c"}},
		{Name: "search", ViewBox: "0 0 24 24", Body: `<path d="M10 0"></path>`, Paths: []string{"M10 0c1 2"}},
	}
}

func TestApplyTextBefore(t *testing.T) {
	rule := &FileRule{Mode: "text", Marker: "</defs>", Position: "before", Template: `<symbol id="{{.Prefix}}{{.Name}}" viewBox="{{.ViewBox}}">{{.Body}}</symbol>`}
	content := []byte("<svg>\n<defs>\n</defs>\n</svg>\n")
	out, n, err := applyText(rule, "ee-icon-", content, testIcons())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 applied, got %d", n)
	}
	s := string(out)
	if !strings.Contains(s, `<symbol id="ee-icon-house-solid" viewBox="0 0 32 32"><path d="M17.363 0.537"></path></symbol>`) {
		t.Errorf("missing first symbol:\n%s", s)
	}
	if !strings.Contains(s, `<symbol id="ee-icon-search" viewBox="0 0 24 24">`) {
		t.Errorf("missing second symbol:\n%s", s)
	}
	if !strings.HasSuffix(s, "</svg>\n") {
		t.Errorf("svg not closed:\n%s", s)
	}

	out2, n2, err := applyText(rule, "ee-icon-", out, testIcons())
	if err != nil {
		t.Fatalf("unexpected error on rerun: %v", err)
	}
	if n2 != 0 {
		t.Errorf("expected idempotent rerun (0 applied), got %d", n2)
	}
	if string(out2) != string(out) {
		t.Errorf("rerun changed content")
	}
}

func TestApplyTextEnd(t *testing.T) {
	rule := &FileRule{Mode: "text", Position: "end", Template: "\n.{{.Prefix}}{{.Name}} {\n  width: 1.75em;\n}"}
	out, n, err := applyText(rule, "ee-icon-", []byte(".ee-icon {}\n"), testIcons())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 applied, got %d", n)
	}
	s := string(out)
	if !strings.Contains(s, "\n.ee-icon-house-solid {\n  width: 1.75em;\n}\n") {
		t.Errorf("missing first rule:\n%s", s)
	}
	if !strings.Contains(s, ".ee-icon-search") {
		t.Errorf("missing second rule:\n%s", s)
	}
}

func TestApplyTextReplace(t *testing.T) {
	rule := &FileRule{Mode: "text", Marker: "<!-- icons -->", Position: "replace", Template: `{{.Name}}`}
	out, n, err := applyText(rule, "ee-icon-", []byte("a<!-- icons -->b"), testIcons())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 applied, got %d", n)
	}
	if string(out) != "ahouse-solid\nsearchb" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestApplyTextMarkerMissing(t *testing.T) {
	rule := &FileRule{Mode: "text", Marker: "zzz", Position: "before", Template: `{{.Name}}`}
	if _, _, err := applyText(rule, "ee-icon-", []byte("abc"), testIcons()); err == nil {
		t.Fatal("expected error for missing marker, got nil")
	}
}

func TestApplyTextIdempotentExisting(t *testing.T) {
	rule := &FileRule{Mode: "text", Marker: "</defs>", Position: "before", Template: `<symbol id="{{.Prefix}}{{.Name}}">{{.Body}}</symbol>`}
	content := []byte(`<svg><defs><symbol id="ee-icon-house-solid"><path d="M17.363 0.537"></path></symbol></defs></svg>`)
	out, n, err := applyText(rule, "ee-icon-", content, testIcons())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected only 1 new icon applied, got %d", n)
	}
	if strings.Contains(string(out), `id="ee-icon-house-solid"`) &&
		strings.Count(string(out), `id="ee-icon-house-solid"`) != 1 {
		t.Errorf("house-solid should appear once: %s", out)
	}
}

func TestApplyIcoMoon(t *testing.T) {
	content := `{
  "metadata": { "name": "EasyEdge", "created": 1637334254503 },
  "iconSets": [
    {
      "selection": [
        { "order": 37, "id": 33, "name": "field-bound-tags-folder", "prevSize": 32 }
      ],
      "id": 0,
      "metadata": { "name": "EasyEdge", "importSize": { "width": 49, "height": 43 } },
      "height": 1024,
      "prevSize": 32,
      "icons": [
        { "id": 33, "paths": ["M1024 800c"], "attrs": [{}], "width": 1012, "isMulticolor": false, "isMulticolor2": false, "grid": 32, "tags": ["field-bound-tags-folder"] }
      ]
    }
  ],
  "uid": -1,
  "preferences": { "fontPref": { "prefix": "ee-icon-" } }
}
`

	out, n, err := applyIcoMoon([]byte(content), testIcons())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 applied, got %d", n)
	}
	s := string(out)

	var doc struct {
		IconSets []struct {
			Selection []icoSelection `json:"selection"`
			Icons     []icoIcon      `json:"icons"`
		} `json:"iconSets"`
		UID         int `json:"uid"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output is not valid json: %v", err)
	}
	if doc.UID != -1 {
		t.Errorf("uid not preserved: %d", doc.UID)
	}
	if !strings.Contains(s, `"name": "EasyEdge"`) {
		t.Errorf("metadata lost: %s", s)
	}
	if !strings.Contains(s, `"prefix": "ee-icon-"`) {
		t.Errorf("preferences lost: %s", s)
	}

	set := doc.IconSets[0]
	if len(set.Icons) != 3 {
		t.Fatalf("expected 3 icons, got %d", len(set.Icons))
	}
	if set.Icons[0].ID != 34 || set.Icons[0].Tags[0] != "house-solid" {
		t.Errorf("unexpected first icon: %+v", set.Icons[0])
	}
	if set.Icons[1].ID != 35 || set.Icons[1].Tags[0] != "search" || len(set.Icons[1].Paths) != 1 {
		t.Errorf("unexpected second icon: %+v", set.Icons[1])
	}
	if set.Selection[0].Order != 38 || set.Selection[0].Name != "house-solid" {
		t.Errorf("unexpected first selection: %+v", set.Selection[0])
	}
	if set.Selection[1].ID != 35 || set.Selection[1].Name != "search" {
		t.Errorf("unexpected second selection: %+v", set.Selection[1])
	}

	out2, n2, err := applyIcoMoon(out, testIcons())
	if err != nil {
		t.Fatalf("unexpected error on rerun: %v", err)
	}
	if n2 != 0 {
		t.Errorf("expected idempotent rerun (0 applied), got %d", n2)
	}
	if string(out2) != string(out) {
		t.Errorf("rerun changed content")
	}
}

func TestApplyIcoMoonInvalidJSON(t *testing.T) {
	if _, _, err := applyIcoMoon([]byte("not json"), testIcons()); err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}

func TestApplyIcoMoonNoIconSets(t *testing.T) {
	if _, _, err := applyIcoMoon([]byte(`{"metadata": {}}`), testIcons()); err == nil {
		t.Fatal("expected error for missing iconSets, got nil")
	}
}

func TestApplyIcoMoonPreservesFormatting(t *testing.T) {
	content := []byte(`{
  "metadata": {
    "name": "EasyEdge",
    "lastOpened": 0,
    "created": 1637334254503
  },
  "iconSets": [
    {
      "selection": [
        {
          "order": 37,
          "id": 33,
          "name": "field-bound-tags-folder",
          "prevSize": 32
        }
      ],
      "id": 0,
      "metadata": { "name": "EasyEdge", "importSize": { "width": 49, "height": 43 } },
      "height": 1024,
      "prevSize": 32,
      "icons": [
        {
          "id": 33,
          "paths": ["M1024 800c"],
          "attrs": [{}],
          "width": 1012,
          "isMulticolor": false,
          "isMulticolor2": false,
          "grid": 32,
          "tags": ["field-bound-tags-folder"]
        }
      ]
    }
  ],
  "uid": -1,
  "preferences": { "fontPref": { "prefix": "ee-icon-" } }
}
`)

	out, n, err := applyIcoMoon(content, testIcons())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 applied, got %d", n)
	}

	selOpen, selClose, err := arrayBounds(content, "selection")
	if err != nil {
		t.Fatal(err)
	}
	icOpen, icClose, err := arrayBounds(content, "icons")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(out[:selOpen], content[:selOpen]) {
		t.Errorf("content before selection changed")
	}
	if !strings.HasSuffix(string(out), string(content[icClose+1:])) {
		t.Errorf("content after icons changed")
	}
	if middle := content[selClose+1 : icOpen]; !strings.Contains(string(out), string(middle)) {
		t.Errorf("content between arrays changed")
	}

	origSel := `        {
          "order": 37,
          "id": 33,
          "name": "field-bound-tags-folder",
          "prevSize": 32
        }`
	if !strings.Contains(string(out), origSel) {
		t.Errorf("original selection entry not preserved verbatim:\n%s", out)
	}
	if strings.Count(string(out), `"order": 37`) != 1 {
		t.Errorf("original selection entry duplicated:\n%s", out)
	}
	if !strings.Contains(string(out), `          "id": 33,`) {
		t.Errorf("original icon entry not preserved verbatim:\n%s", out)
	}

	var doc struct {
		IconSets []struct {
			Selection []icoSelection `json:"selection"`
			Icons     []icoIcon      `json:"icons"`
		} `json:"iconSets"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output is not valid json: %v", err)
	}
	if doc.IconSets[0].Selection[0].ID != 34 {
		t.Errorf("expected new selection first, got %+v", doc.IconSets[0].Selection[0])
	}
	if doc.IconSets[0].Icons[0].ID != 34 {
		t.Errorf("expected new icon first, got %+v", doc.IconSets[0].Icons[0])
	}
}

func TestApplyIconsLocalProject(t *testing.T) {
	base := t.TempDir()
	svg := filepath.Join(base, "sprite.svg")
	if err := os.WriteFile(svg, []byte("<svg>\n<defs>\n</defs>\n</svg>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	jsonFile := filepath.Join(base, "project.json")
	if err := os.WriteFile(jsonFile, []byte(`{
  "metadata": { "name": "EasyEdge" },
  "iconSets": [ { "selection": [], "id": 0, "metadata": {}, "height": 1024, "prevSize": 32, "icons": [] } ],
  "uid": -1,
  "preferences": {}
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	css := filepath.Join(base, "style.css")
	if err := os.WriteFile(css, []byte(".ee-icon {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		ProjectPath: base,
		IconPrefix:  "ee-icon-",
		Files: []FileRule{
			{Name: "sprite.svg", Mode: "text", Marker: "</defs>", Position: "before", Template: `<symbol id="{{.Prefix}}{{.Name}}" viewBox="{{.ViewBox}}">{{.Body}}</symbol>`},
			{Name: "project.json", Mode: "icomoon"},
			{Name: "style.css", Mode: "text", Position: "end", Template: "\n.{{.Prefix}}{{.Name}} { width: 1.75em; }"},
		},
	}

	if err := applyIcons(newProject(base), cfg, testIcons()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svgData, _ := os.ReadFile(svg)
	if !strings.Contains(string(svgData), `id="ee-icon-house-solid"`) {
		t.Errorf("svg not updated: %s", svgData)
	}
	jsonData, _ := os.ReadFile(jsonFile)
	if !strings.Contains(string(jsonData), `"search"`) {
		t.Errorf("json not updated: %s", jsonData)
	}
	cssData, _ := os.ReadFile(css)
	if !strings.Contains(string(cssData), ".ee-icon-search") {
		t.Errorf("css not updated: %s", cssData)
	}
	if _, err := os.Stat(svg + ".orig"); !os.IsNotExist(err) {
		t.Errorf("backup should be removed after success: %v", err)
	}
	if _, err := os.Stat(jsonFile + ".orig"); !os.IsNotExist(err) {
		t.Errorf("json backup should be removed after success: %v", err)
	}

	if err := applyIcons(newProject(base), cfg, testIcons()); err != nil {
		t.Fatalf("unexpected error on rerun: %v", err)
	}
	svgData, _ = os.ReadFile(svg)
	if strings.Count(string(svgData), `id="ee-icon-house-solid"`) != 1 {
		t.Errorf("rerun duplicated icons: %s", svgData)
	}
	if _, err := os.Stat(svg + ".orig"); !os.IsNotExist(err) {
		t.Errorf("backup should not exist after rerun: %v", err)
	}
}

func TestApplyIconsKeepsBackupOnError(t *testing.T) {
	base := t.TempDir()
	svg := filepath.Join(base, "sprite.svg")
	if err := os.WriteFile(svg, []byte("<svg>\n<defs>\n</defs>\n</svg>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	css := filepath.Join(base, "style.css")
	if err := os.WriteFile(css, []byte(".ee-icon {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		ProjectPath: base,
		IconPrefix:  "ee-icon-",
		Files: []FileRule{
			{Name: "sprite.svg", Mode: "text", Marker: "</defs>", Position: "before", Template: `<symbol id="{{.Prefix}}{{.Name}}">{{.Body}}</symbol>`},
			{Name: "style.css", Mode: "text", Marker: "zzz-missing", Position: "before", Template: `{{.Name}}`},
		},
	}
	if err := applyIcons(newProject(base), cfg, testIcons()); err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, err := os.Stat(svg + ".orig"); err != nil {
		t.Errorf("backup of the first file should remain after error: %v", err)
	}
	if _, err := os.Stat(css + ".orig"); err == nil {
		t.Errorf("no backup expected for the failing file: %v", err)
	}
}
