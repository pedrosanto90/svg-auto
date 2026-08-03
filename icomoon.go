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

const (
	fileInputSelector    = "input[type=file]"
	selectedIconSelector = "#set0 [class~='mi-selected']"
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
	return fmt.Errorf("%s (timeout after %s)", errMsg, timeout)
}

func importIcons(ctx context.Context, files []string) error {
	if err := dismissWelcome(ctx); err != nil {
		return err
	}

	ready := `document.querySelector('#set0') !== null || document.querySelector('input[type=file]') !== null`
	if err := waitForJS(ctx, ready, settleTimeout, "the IcoMoon app did not finish loading"); err != nil {
		return err
	}

	var hasDefaultSet bool
	_ = evalJS(ctx, `!!document.querySelector('#set0')`, &hasDefaultSet)
	if hasDefaultSet {
		if err := clickSetMenuItem(ctx, "Remove Set"); err != nil {
			return fmt.Errorf("could not remove the default IcoMoon set: %w", err)
		}
		if err := waitForJS(ctx, `!document.querySelector('#set0')`, importTimeout, "the default set was not removed"); err != nil {
			return err
		}
	}

	if err := waitForJS(ctx, `(() => { const input = document.querySelector('input[type=file]'); return !!input && !input.disabled; })()`, importTimeout, "the import area did not appear"); err != nil {
		return err
	}
	if err := chromedp.Run(ctx, chromedp.SetUploadFiles(fileInputSelector, files)); err != nil {
		return fmt.Errorf("failed to upload files to IcoMoon: %w", err)
	}

	cond := fmt.Sprintf(`(() => { const s = document.querySelector('#set0'); return !!s && s.querySelectorAll("[class~='miBox']").length >= %d; })()`, len(files))
	if err := waitForJS(ctx, cond, importTimeout, fmt.Sprintf("the %d icons were not imported", len(files))); err != nil {
		return err
	}

	if err := selectAllIcons(ctx, len(files)); err != nil {
		return err
	}
	return nil
}

func dismissWelcome(ctx context.Context) error {
	code := `(() => {
		const visible = e => e && e.offsetParent !== null && !e.disabled;
		const label = e => ((e.getAttribute('aria-label') || '') + ' ' + (e.textContent || '')).trim().toLowerCase();
		const d = Array.from(document.querySelectorAll('[role=dialog] button, .overlay button, .modal button'))
			.find(b => visible(b) && label(b).includes('dismiss'));
		if (d) d.click();
		return true;
	})()`
	return evalJS(ctx, code, new(bool))
}

func clickSetMenu(ctx context.Context) error {
	code := `(() => {
		const visible = e => e && e.offsetParent !== null && !e.disabled;
		const label = e => ((e.getAttribute('aria-label') || '') + ' ' + (e.textContent || '')).trim().toLowerCase();
		const m = Array.from(document.querySelectorAll('[role=menu], .menuList2')).find(visible);
		if (m) return true;
		const bs = Array.from(document.querySelectorAll('button, [role=button]'))
			.filter(b => visible(b) && label(b) === 'menu');
		if (bs.length !== 1) return false;
		bs[0].click();
		return true;
	})()`
	return waitForJS(ctx, code, selectTimeout, "set menu button not found")
}

func clickSetMenuItem(ctx context.Context, text string) error {
	ready := fmt.Sprintf(`Array.from(document.querySelectorAll('[role=menu] li, .menuList2 li')).some(li => (li.textContent||'').includes(%q))`, text)
	if err := waitForJS(ctx, ready, selectTimeout, fmt.Sprintf("menu item %q unavailable", text)); err != nil {
		return err
	}
	if err := clickSetMenu(ctx); err != nil {
		return err
	}

	deadline := time.Now().Add(selectTimeout)
	for time.Now().Before(deadline) {
		code := fmt.Sprintf(`(() => {
			const t = %q;
			const visible = e => e && e.offsetParent !== null && !e.disabled;
			const li = Array.from(document.querySelectorAll('[role=menu] li, .menuList2 li'))
				.find(li => visible(li) && (li.textContent||'').includes(t));
			if (!li) return false;
			const b = li.querySelector('button, [role=button]');
			if (b && !visible(b)) return false;
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
	return fmt.Errorf("menu item %q not found", text)
}

func selectAllIcons(ctx context.Context, n int) error {
	cond := fmt.Sprintf(`document.querySelectorAll("%s").length === %d`, selectedIconSelector, n)
	var selected int
	_ = evalJS(ctx, fmt.Sprintf(`document.querySelectorAll("%s").length`, selectedIconSelector), &selected)
	if selected == n {
		return nil
	}
	if err := clickSetMenuItem(ctx, "Select All"); err != nil {
		return err
	}
	return waitForJS(ctx, cond, selectTimeout, "could not select all icons")
}

func downloadIcons(ctx context.Context, downloadDir string) (string, error) {
	if err := chromedp.Run(ctx,
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).WithDownloadPath(downloadDir).WithEventsEnabled(true),
		chromedp.Navigate(icomoonImageURL),
	); err != nil {
		return "", fmt.Errorf("failed to open the export page: %w", err)
	}

	dlBtn := `(() => {
		const visible = e => e && e.offsetParent !== null && !e.disabled;
		const label = e => ((e.getAttribute('aria-label') || '') + ' ' + (e.textContent || '')).trim().toLowerCase();
		return Array.from(document.querySelectorAll('button, [role=button]'))
			.some(b => visible(b) && label(b).includes('download'));
	})()`
	if err := waitForJS(ctx, dlBtn, downloadTimeout, "download button not found"); err != nil {
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
		const visible = e => e && e.offsetParent !== null && !e.disabled;
		const label = e => ((e.getAttribute('aria-label') || '') + ' ' + (e.textContent || '')).trim().toLowerCase();
		const bs = Array.from(document.querySelectorAll('button, [role=button]'))
			.filter(b => visible(b) && label(b).includes('download'));
		const b = bs.length === 1 ? bs[0] : null;
		if (!b) return false;
		b.click();
		return true;
	})()`
	if err := evalJS(ctx, code, &clicked); err != nil {
		return "", err
	}
	if !clicked {
		return "", fmt.Errorf("could not click the download button")
	}

	var guid string
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("timeout waiting for the download: %w", ctx.Err())
	case guid = <-done:
	}

	zipPath := filepath.Join(downloadDir, guid)
	if err := waitForFileStable(ctx, zipPath, downloadTimeout); err != nil {
		return "", err
	}
	if !isZipFile(zipPath) {
		return "", fmt.Errorf("the downloaded file %s is not a valid zip", zipPath)
	}

	name := suggestedName
	if strings.TrimSpace(name) == "" {
		name = "icons.zip"
	}
	finalPath := filepath.Join(downloadDir, filepath.Base(name))
	if err := os.Rename(zipPath, finalPath); err != nil {
		return "", fmt.Errorf("failed to name the downloaded file: %w", err)
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
			return fmt.Errorf("file %s did not finish downloading: %w", path, ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("file %s did not finish downloading (timeout after %s)", path, timeout)
}
