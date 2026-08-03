package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
)

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

func extractSymbols(svg []byte) ([]string, error) {
	re := regexp.MustCompile(`(?s)<symbol\b[^>]*>.*?</symbol>`)
	symbols := re.FindAllString(string(svg), -1)
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols found in symbol-defs.svg")
	}
	return symbols, nil
}

func extractPaths(data []byte) ([]string, error) {
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

	var paths []string
	for _, ic := range doc.Icons {
		paths = append(paths, ic.Icon.Paths...)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no paths found in selection.json")
	}
	return paths, nil
}

func processZip(zipPath string) (symbols []string, paths []string, err error) {
	svg, err := readZipEntry(zipPath, "symbol-defs.svg")
	if err != nil {
		return nil, nil, err
	}
	sel, err := readZipEntry(zipPath, "selection.json")
	if err != nil {
		return nil, nil, err
	}

	symbols, err = extractSymbols(svg)
	if err != nil {
		return nil, nil, err
	}
	paths, err = extractPaths(sel)
	if err != nil {
		return nil, nil, err
	}
	return symbols, paths, nil
}
