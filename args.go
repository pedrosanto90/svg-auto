package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const usage = `Uso: svg-auto <ficheiro1.svg> [ficheiro2.svg ...]

Importa ficheiros SVG para o IcoMoon e descarrega os SVGs processados para ./output/svg.

Opções:
  -h, --help    mostra esta ajuda

Variáveis de ambiente:
  SVG_AUTO_BROWSER    caminho ou nome do executável do browser (opcional)`

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
			return nil, fmt.Errorf("falha ao obter caminho absoluto de %q: %w", arg, err)
		}
		files = append(files, abs)
	}
	return files, nil
}

func validateSVG(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("ficheiro não encontrado: %q", path)
		}
		return fmt.Errorf("falha ao aceder a %q: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("esperado um ficheiro, mas %q é uma pasta", path)
	}
	if !strings.EqualFold(filepath.Ext(path), ".svg") {
		return fmt.Errorf("%q não termina em .svg", path)
	}
	return nil
}
