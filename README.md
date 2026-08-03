# svg-auto

Automation for [IcoMoon](https://icomoon.io) to import SVG files and download the generated package, using [chromedp](https://github.com/chromedp/chromedp).

## Installation

### Option 1: go install (recommended)

Requires Go 1.26+.

```sh
go install github.com/pedrosanto90/svg-auto@latest
```

This installs the `svg-auto` binary into `$GOBIN` (by default `$GOPATH/bin`). Make sure that directory is in your `PATH`:

```sh
export PATH=$PATH:$(go env GOPATH)/bin
```

If you have a local checkout, you can install from source:

```sh
go install .
```

To update to the latest version:

```sh
go install github.com/pedrosanto90/svg-auto@latest
```

### Option 2: System-wide install

Install the binary to `/usr/local/bin` using the Makefile:

```sh
make build
sudo make install
```

Or manually:

```sh
go build -o svg-auto .
sudo install -m 0755 svg-auto /usr/local/bin/svg-auto
```

To remove it:

```sh
sudo make uninstall
```

### Verify the installation

```sh
svg-auto -h
```

> **Note:** `go install github.com/pedrosanto90/svg-auto@latest` requires the repository to be up to date on GitHub (commits must be pushed).

## Supported browsers

Works with any Chromium-based browser. Automatic detection looks for, in order:

- `brave-browser`, `brave-browser-stable`
- `chromium`, `chromium-browser`
- `google-chrome`, `google-chrome-stable`
- `microsoft-edge`
- `vivaldi`, `opera`, `chrome`

## How to run

```sh
svg-auto icon1.svg icon2.svg icon3.svg
```

Or from a local checkout:

```sh
go run . icon1.svg icon2.svg icon3.svg
```

The script imports the SVG files into IcoMoon and downloads the generated package (a `.zip` file, with the original IcoMoon name) to `./output/`. The package contains the `SVG/` and `PNG/` folders, `selection.json`, and other generated files.

To use a specific browser (path or executable name):

```sh
SVG_AUTO_BROWSER=brave-browser svg-auto icon.svg
```

To view the help:

```sh
svg-auto -h
```

## Requirements

- Go 1.26+
- A Chromium-based browser installed (or set via `SVG_AUTO_BROWSER`)
