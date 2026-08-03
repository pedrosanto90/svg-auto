package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const browserEnvVar = "SVG_AUTO_BROWSER"

var browserNames = []string{
	"brave-browser",
	"brave-browser-stable",
	"chromium",
	"chromium-browser",
	"google-chrome",
	"google-chrome-stable",
	"microsoft-edge",
	"vivaldi",
	"opera",
	"chrome",
}

func findBrowser() (string, error) {
	if override := os.Getenv(browserEnvVar); override != "" {
		path, err := exec.LookPath(override)
		if err != nil {
			return "", fmt.Errorf("não foi possível localizar o browser definido em %s: %q (%w)", browserEnvVar, override, err)
		}
		return path, nil
	}

	searched := make([]string, 0, len(browserNames))
	for _, name := range browserNames {
		searched = append(searched, name)
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	for _, path := range absBrowserPaths() {
		searched = append(searched, path)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("não foi encontrado nenhum browser suportado. Procurei por: %s. Instale um browser Chromium-based ou defina %s com o caminho do executável",
		strings.Join(searched, ", "), browserEnvVar)
}

func absBrowserPaths() []string {
	var paths []string
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths,
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		)
	case "windows":
		paths = append(paths,
			filepath.Join(os.Getenv("PROGRAMFILES"), "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
			filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
			filepath.Join(os.Getenv("PROGRAMFILES"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("PROGRAMFILES"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
		)
	default:
		paths = append(paths,
			"/usr/bin/brave-browser",
			"/usr/bin/brave-browser-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/usr/bin/google-chrome",
			"/snap/bin/chromium",
		)
	}
	return paths
}
