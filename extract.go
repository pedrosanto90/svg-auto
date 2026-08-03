package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
)

type Icon struct {
	Name    string
	ViewBox string
	Body    string
	Paths   []string
}

func readZipEntry(zipPath, name string) ([]byte, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", zipPath, err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if filepath.Base(f.Name) == name {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open %s inside %s: %w", name, zipPath, err)
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("file %s not found in zip %s", name, zipPath)
}

var (
	symbolRe = regexp.MustCompile(`(?s)<symbol\b[^>]*>.*?</symbol>`)
	idRe     = regexp.MustCompile(`\bid="([^"]*)"`)
	viewBoxRe = regexp.MustCompile(`\bviewBox="([^"]*)"`)
)

func extractSymbols(svg []byte) ([]Icon, error) {
	matches := symbolRe.FindAllString(string(svg), -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no symbols found in symbol-defs.svg")
	}

	icons := make([]Icon, 0, len(matches))
	for _, m := range matches {
		open := strings.Index(m, ">")
		if open < 0 {
			continue
		}
		head := m[:open]
		body := strings.TrimSpace(m[open+1 : len(m)-len("</symbol>")])

		ic := Icon{Body: body}
		if sm := idRe.FindStringSubmatch(head); sm != nil {
			ic.Name = strings.TrimPrefix(sm[1], "icon-")
		}
		if sm := viewBoxRe.FindStringSubmatch(head); sm != nil {
			ic.ViewBox = sm[1]
		}
		icons = append(icons, ic)
	}
	if len(icons) == 0 {
		return nil, fmt.Errorf("no valid symbols found in symbol-defs.svg")
	}
	return icons, nil
}

func extractPaths(data []byte) ([][]string, error) {
	var doc struct {
		Icons []struct {
			Icon struct {
				Paths []string `json:"paths"`
			} `json:"icon"`
		} `json:"icons"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse selection.json: %w", err)
	}

	paths := make([][]string, 0, len(doc.Icons))
	for _, ic := range doc.Icons {
		if len(ic.Icon.Paths) == 0 {
			return nil, fmt.Errorf("no paths found in selection.json")
		}
		paths = append(paths, ic.Icon.Paths)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no paths found in selection.json")
	}
	return paths, nil
}

func processZip(zipPath string) ([]Icon, error) {
	svg, err := readZipEntry(zipPath, "symbol-defs.svg")
	if err != nil {
		return nil, err
	}
	sel, err := readZipEntry(zipPath, "selection.json")
	if err != nil {
		return nil, err
	}

	symbols, err := extractSymbols(svg)
	if err != nil {
		return nil, err
	}
	paths, err := extractPaths(sel)
	if err != nil {
		return nil, err
	}

	for i := range symbols {
		if i < len(paths) {
			symbols[i].Paths = paths[i]
		}
	}
	return symbols, nil
}
