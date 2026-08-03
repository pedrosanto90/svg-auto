package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const usage = `Usage: svg-auto <file1.svg> [file2.svg ...]

Imports SVG files into IcoMoon and downloads the generated package (.zip) to ./output/.

Options:
  -h, --help    show this help

Environment variables:
  SVG_AUTO_BROWSER    path or name of the browser executable (optional)`

func parseArgs() ([]string, error) {
	args := os.Args[1:]

	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println(usage)
			os.Exit(0)
		}
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("%s", usage)
	}

	files := make([]string, 0, len(args))
	for _, arg := range args {
		if err := validateSVG(arg); err != nil {
			return nil, err
		}
		abs, err := filepath.Abs(arg)
		if err != nil {
			return nil, fmt.Errorf("failed to get the absolute path of %q: %w", arg, err)
		}
		files = append(files, abs)
	}
	return files, nil
}

func validateSVG(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %q", path)
		}
		return fmt.Errorf("failed to access %q: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected a file, but %q is a directory", path)
	}
	if !strings.EqualFold(filepath.Ext(path), ".svg") {
		return fmt.Errorf("%q does not end in .svg", path)
	}
	return nil
}
