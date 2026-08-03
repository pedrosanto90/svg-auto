package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

type iconData struct {
	Prefix    string
	Name      string
	ViewBox   string
	Body      string
	Paths     string
	PathsJson string
	TagsJson  string
}

type icoSelection struct {
	Order    int    `json:"order"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	PrevSize int    `json:"prevSize"`
}

type icoIcon struct {
	ID            int      `json:"id"`
	Paths         []string `json:"paths"`
	Attrs         []any    `json:"attrs"`
	Width         int      `json:"width"`
	IsMulticolor  bool     `json:"isMulticolor"`
	IsMulticolor2 bool     `json:"isMulticolor2"`
	Grid          int      `json:"grid"`
	Tags          []string `json:"tags"`
}

type icoSet struct {
	Selection []icoSelection  `json:"selection"`
	ID        int             `json:"id"`
	Metadata  json.RawMessage `json:"metadata"`
	Height    int             `json:"height"`
	PrevSize  int             `json:"prevSize"`
	Icons     []icoIcon       `json:"icons"`
}

type icoProject struct {
	Metadata    json.RawMessage `json:"metadata"`
	IconSets    []icoSet        `json:"iconSets"`
	UID         int             `json:"uid"`
	Preferences json.RawMessage `json:"preferences"`
}

func applyIcons(p *project, cfg *Config, icons []Icon) error {
	for _, rule := range cfg.Files {
		if rule.Mode == "icomoon" {
			for _, ic := range icons {
				if len(ic.Paths) == 0 {
					return fmt.Errorf("icon %q has no paths (required for %s)", ic.Name, rule.Name)
				}
			}
		}
	}

	var backups []string
	for i := range cfg.Files {
		rule := &cfg.Files[i]

		content, err := p.readFile(rule.Name)
		if err != nil {
			return err
		}

		var updated []byte
		var applied int
		switch rule.Mode {
		case "text":
			updated, applied, err = applyText(rule, cfg.IconPrefix, content, icons)
		case "icomoon":
			updated, applied, err = applyIcoMoon(content, icons)
		}
		if err != nil {
			return fmt.Errorf("failed to edit %s: %w", rule.Name, err)
		}

		if applied == 0 {
			fmt.Printf("No changes for %s.\n", rule.Name)
			continue
		}
		if err := p.backup(rule.Name); err != nil {
			return err
		}
		backups = append(backups, rule.Name+".orig")
		if err := p.writeFile(rule.Name, updated); err != nil {
			return err
		}
		fmt.Printf("Applied %d icon(s) to %s.\n", applied, rule.Name)
	}

	for _, b := range backups {
		if err := p.removeFile(b); err != nil {
			return err
		}
	}
	return nil
}

func applyText(rule *FileRule, prefix string, content []byte, icons []Icon) ([]byte, int, error) {
	var blocks []string
	for _, ic := range icons {
		blk, err := renderIcon(rule, prefix, ic)
		if err != nil {
			return nil, 0, err
		}
		if bytes.Contains(content, blk) {
			continue
		}
		blocks = append(blocks, strings.TrimSpace(string(blk)))
	}
	if len(blocks) == 0 {
		return content, 0, nil
	}

	inserted := strings.Join(blocks, "\n")
	marker := []byte(rule.Marker)

	var b bytes.Buffer
	switch rule.Position {
	case "end":
		b.Write(content)
		if len(content) > 0 && content[len(content)-1] != '\n' {
			b.WriteByte('\n')
		}
		b.WriteString(inserted)
		b.WriteByte('\n')
	case "replace":
		idx := bytes.Index(content, marker)
		if idx < 0 {
			return nil, 0, fmt.Errorf("marker %q not found", rule.Marker)
		}
		b.Write(content[:idx])
		b.WriteString(inserted)
		b.Write(content[idx+len(marker):])
	case "before":
		idx := bytes.Index(content, marker)
		if idx < 0 {
			return nil, 0, fmt.Errorf("marker %q not found", rule.Marker)
		}
		b.Write(content[:idx])
		b.WriteString(inserted)
		b.WriteByte('\n')
		b.Write(content[idx:])
	case "after":
		idx := bytes.Index(content, marker)
		if idx < 0 {
			return nil, 0, fmt.Errorf("marker %q not found", rule.Marker)
		}
		idx += len(marker)
		b.Write(content[:idx])
		b.WriteByte('\n')
		b.WriteString(inserted)
		b.Write(content[idx:])
	default:
		return nil, 0, fmt.Errorf("invalid position %q", rule.Position)
	}
	return b.Bytes(), len(blocks), nil
}

func renderIcon(rule *FileRule, prefix string, ic Icon) ([]byte, error) {
	pathsJSON, err := json.Marshal(ic.Paths)
	if err != nil {
		return nil, err
	}
	tagsJSON, err := json.Marshal([]string{ic.Name})
	if err != nil {
		return nil, err
	}

	data := iconData{
		Prefix:    prefix,
		Name:      ic.Name,
		ViewBox:   ic.ViewBox,
		Body:      strings.TrimSpace(ic.Body),
		Paths:     strings.Join(ic.Paths, ","),
		PathsJson: string(pathsJSON),
		TagsJson:  string(tagsJSON),
	}

	tmpl, err := template.New("rule").Parse(rule.Template)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	return buf.Bytes(), nil
}

func applyIcoMoon(content []byte, icons []Icon) ([]byte, int, error) {
	var doc icoProject
	if err := json.Unmarshal(content, &doc); err != nil {
		return nil, 0, fmt.Errorf("failed to parse JSON: %w", err)
	}
	if len(doc.IconSets) == 0 {
		return nil, 0, fmt.Errorf("no iconSets found in JSON")
	}
	set := &doc.IconSets[0]

	maxID := 0
	for _, ic := range set.Icons {
		if ic.ID > maxID {
			maxID = ic.ID
		}
	}
	maxOrder := 0
	for _, s := range set.Selection {
		if s.Order > maxOrder {
			maxOrder = s.Order
		}
	}

	existing := make(map[string]bool)
	for _, ic := range set.Icons {
		for _, t := range ic.Tags {
			existing[t] = true
		}
	}

	var newSel []icoSelection
	var newIcons []icoIcon
	for _, ic := range icons {
		if existing[ic.Name] {
			continue
		}
		maxID++
		maxOrder++
		existing[ic.Name] = true

		newSel = append(newSel, icoSelection{
			Order:    maxOrder,
			ID:       maxID,
			Name:     ic.Name,
			PrevSize: 32,
		})
		newIcons = append(newIcons, icoIcon{
			ID:            maxID,
			Paths:         ic.Paths,
			Attrs:         []any{map[string]any{}},
			Width:         1024,
			IsMulticolor:  false,
			IsMulticolor2: false,
			Grid:          32,
			Tags:          []string{ic.Name},
		})
	}
	if len(newIcons) == 0 {
		return content, 0, nil
	}

	set.Selection = append(newSel, set.Selection...)
	set.Icons = append(newIcons, set.Icons...)

	out, err := json.MarshalIndent(&doc, "", "  ")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to serialize JSON: %w", err)
	}
	return append(out, '\n'), len(newIcons), nil
}
