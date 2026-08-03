package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

const outputDir = "output"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	files, err := parseArgs()
	if err != nil {
		return err
	}

	browserPath, err := findBrowser()
	if err != nil {
		return err
	}
	fmt.Printf("Usando browser: %s\n", browserPath)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browserPath),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	ctx, cancelTimeout := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelTimeout()

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("falha ao criar %s: %w", outputDir, err)
	}
	downloadDir := filepath.Join(outputDir, "download")
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		return fmt.Errorf("falha ao criar %s: %w", downloadDir, err)
	}
	defer os.RemoveAll(downloadDir)

	if err := chromedp.Run(ctx, chromedp.Navigate(icomoonSelectURL)); err != nil {
		return fmt.Errorf("falha ao abrir o IcoMoon: %w", err)
	}
	if err := chromedp.Run(ctx, chromedp.WaitVisible(".w-main", chromedp.ByQuery)); err != nil {
		return fmt.Errorf("falha ao carregar o IcoMoon: %w", err)
	}

	if err := importIcons(ctx, files); err != nil {
		return err
	}
	fmt.Printf("Importados %d ícones.\n", len(files))

	zipPath, err := downloadIcons(ctx, downloadDir)
	if err != nil {
		return err
	}

	finalZip := filepath.Join(outputDir, filepath.Base(zipPath))
	if err := os.Rename(zipPath, finalZip); err != nil {
		return fmt.Errorf("falha ao mover o zip para %s: %w", finalZip, err)
	}

	abs, _ := filepath.Abs(finalZip)
	fmt.Printf("Concluído. Pacote descarregado em %s\n", abs)
	return nil
}
