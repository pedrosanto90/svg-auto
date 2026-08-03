package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

const (
	icomoonSelectURL = "https://icomoon.io/app/#/select"
	icomoonImageURL  = "https://icomoon.io/app/#/select/image"

	settleTimeout   = 30 * time.Second
	importTimeout   = 20 * time.Second
	selectTimeout   = 10 * time.Second
	downloadTimeout = 60 * time.Second
)

func evalJS(ctx context.Context, code string, out any) error {
	return chromedp.Run(ctx, chromedp.Evaluate(code, out))
}

func waitForJS(ctx context.Context, code string, timeout time.Duration, errMsg string) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var ok bool
		if err := evalJS(ctx, code, &ok); err != nil {
			return fmt.Errorf("%s: %w", errMsg, err)
		}
		if ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s: %w", errMsg, ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("%s (timeout após %s)", errMsg, timeout)
}

func importIcons(ctx context.Context, files []string) error {
	if err := dismissWelcome(ctx); err != nil {
		return err
	}

	ready := `document.querySelector('#set0') !== null || document.querySelector('#file') !== null`
	if err := waitForJS(ctx, ready, settleTimeout, "a app do IcoMoon não terminou de carregar"); err != nil {
		return err
	}

	var hasDefaultSet bool
	_ = evalJS(ctx, `!!document.querySelector('#set0')`, &hasDefaultSet)
	if hasDefaultSet {
		if err := clickSetMenuItem(ctx, "Remove Set"); err != nil {
			return fmt.Errorf("não foi possível remover o set predefinido do IcoMoon: %w", err)
		}
		if err := waitForJS(ctx, `!document.querySelector('#set0')`, importTimeout, "o set predefinido não foi removido"); err != nil {
			return err
		}
	}

	if err := waitForJS(ctx, `document.querySelector('#file input[type=file]') !== null`, importTimeout, "a área de importação não apareceu"); err != nil {
		return err
	}
	if err := chromedp.Run(ctx, chromedp.SetUploadFiles("#file input[type=file]", files)); err != nil {
		return fmt.Errorf("falha ao enviar ficheiros para o IcoMoon: %w", err)
	}

	cond := fmt.Sprintf(`(() => { const s = document.querySelector('#set0'); return !!s && s.querySelectorAll('.miBox').length >= %d; })()`, len(files))
	if err := waitForJS(ctx, cond, importTimeout, fmt.Sprintf("os %d ícones não foram importados", len(files))); err != nil {
		return err
	}

	if err := selectAllIcons(ctx, len(files)); err != nil {
		return err
	}
	return nil
}

func dismissWelcome(ctx context.Context) error {
	code := `(() => {
		const d = Array.from(document.querySelectorAll('.overlay button, .modal button')).find(b => (b.textContent||'').includes('Dismiss'));
		if (d) d.click();
		return true;
	})()`
	return evalJS(ctx, code, new(bool))
}

func clickSetMenu(ctx context.Context) error {
	code := `(() => {
		const m = document.querySelector('.menuList2');
		if (m && !m.classList.contains('hidden')) return true;
		const bs = Array.from(document.querySelectorAll('button')).filter(b => (b.textContent||'').trim() === 'Menu');
		if (!bs.length) return false;
		bs[bs.length-1].click();
		return true;
	})()`
	return waitForJS(ctx, code, selectTimeout, "botão de menu do set não encontrado")
}

func clickSetMenuItem(ctx context.Context, text string) error {
	ready := fmt.Sprintf(`Array.from(document.querySelectorAll('.menuList2 li')).some(li => (li.textContent||'').includes(%q))`, text)
	if err := waitForJS(ctx, ready, selectTimeout, fmt.Sprintf("item de menu %q indisponível", text)); err != nil {
		return err
	}
	if err := clickSetMenu(ctx); err != nil {
		return err
	}

	deadline := time.Now().Add(selectTimeout)
	for time.Now().Before(deadline) {
		code := fmt.Sprintf(`(() => {
			const t = %q;
			const li = Array.from(document.querySelectorAll('.menuList2 li')).find(li => li.offsetParent !== null && (li.textContent||'').includes(t));
			if (!li) return false;
			const b = li.querySelector('button');
			if (b) b.click(); else li.click();
			return true;
		})()`, text)
		var clicked bool
		if err := evalJS(ctx, code, &clicked); err != nil {
			return err
		}
		if clicked {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(400 * time.Millisecond):
		}
	}
	return fmt.Errorf("item de menu %q não encontrado", text)
}

func selectAllIcons(ctx context.Context, n int) error {
	cond := fmt.Sprintf(`document.querySelectorAll('#set0 .miBox.mi-selected').length === %d`, n)
	var selected int
	_ = evalJS(ctx, `document.querySelectorAll('#set0 .miBox.mi-selected').length`, &selected)
	if selected == n {
		return nil
	}
	if err := clickSetMenuItem(ctx, "Select All"); err != nil {
		return err
	}
	return waitForJS(ctx, cond, selectTimeout, "não foi possível selecionar todos os ícones")
}

func downloadIcons(ctx context.Context, downloadDir string) (string, error) {
	if err := chromedp.Run(ctx,
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).WithDownloadPath(downloadDir).WithEventsEnabled(true),
		chromedp.Navigate(icomoonImageURL),
	); err != nil {
		return "", fmt.Errorf("falha ao abrir a página de exportação: %w", err)
	}

	dlBtn := `Array.from(document.querySelectorAll('button')).some(b => (b.className||'').includes('btn4') && (b.textContent||'').includes('Download'))`
	if err := waitForJS(ctx, dlBtn, downloadTimeout, "botão de download não encontrado"); err != nil {
		return "", err
	}

	var suggestedName string
	done := make(chan string, 1)
	chromedp.ListenTarget(ctx, func(v any) {
		switch ev := v.(type) {
		case *browser.EventDownloadWillBegin:
			suggestedName = ev.SuggestedFilename
		case *browser.EventDownloadProgress:
			if ev.State == browser.DownloadProgressStateCompleted {
				done <- ev.GUID
			}
		}
	})

	var clicked bool
	code := `(() => {
		const b = Array.from(document.querySelectorAll('button')).find(b => (b.className||'').includes('btn4') && (b.textContent||'').includes('Download'));
		if (!b) return false;
		b.click();
		return true;
	})()`
	if err := evalJS(ctx, code, &clicked); err != nil {
		return "", err
	}
	if !clicked {
		return "", fmt.Errorf("não foi possível clicar no botão de download")
	}

	var guid string
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("timeout à espera do download: %w", ctx.Err())
	case guid = <-done:
	}

	zipPath := filepath.Join(downloadDir, guid)
	if err := waitForFileStable(ctx, zipPath, downloadTimeout); err != nil {
		return "", err
	}
	if !isZipFile(zipPath) {
		return "", fmt.Errorf("o ficheiro descarregado %s não é um zip válido", zipPath)
	}

	name := suggestedName
	if strings.TrimSpace(name) == "" {
		name = "icons.zip"
	}
	finalPath := filepath.Join(downloadDir, filepath.Base(name))
	if err := os.Rename(zipPath, finalPath); err != nil {
		return "", fmt.Errorf("falha ao dar nome ao ficheiro descarregado: %w", err)
	}
	return finalPath, nil
}

func isZipFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	magic := make([]byte, 2)
	if _, err := io.ReadFull(f, magic); err != nil {
		return false
	}
	return magic[0] == 'P' && magic[1] == 'K'
}

func waitForFileStable(ctx context.Context, path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var prevSize int64
	for time.Now().Before(deadline) {
		fi, err := os.Stat(path)
		if err == nil {
			if fi.Size() > 0 && fi.Size() == prevSize {
				return nil
			}
			if fi.Size() > 0 {
				prevSize = fi.Size()
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("ficheiro %s não concluiu o download: %w", path, ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("ficheiro %s não concluiu o download (timeout após %s)", path, timeout)
}
