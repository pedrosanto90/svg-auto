# svg-auto

Automation for [IcoMoon](https://icomoon.io) to import SVG files and download the generated package, using [chromedp](https://github.com/chromedp/chromedp).

## Installation

### Prerequisites

- [Go](https://go.dev/dl/) 1.26+
- A Chromium-based browser installed (Brave, Chrome, Chromium, Microsoft Edge, etc.)

### Option 1: Clone and build from source

Clone the repository:

```sh
git clone https://github.com/pedrosanto90/svg-auto.git
cd svg-auto
```

Build the binary:

```sh
go build -o svg-auto .
```

Run it directly from the checkout:

```sh
./svg-auto -h
```

Or install it system-wide with the Makefile:

```sh
sudo make install
```

### Option 2: go install

Install directly with the Go toolchain. Requires Go 1.26+.

```sh
go install github.com/pedrosanto90/svg-auto@latest
```

This installs the `svg-auto` binary into `$GOBIN` (by default `$GOPATH/bin`). Make sure that directory is in your `PATH`:

```sh
export PATH=$PATH:$(go env GOPATH)/bin
```

### Verify the installation

```sh
svg-auto -h
```

You should see the usage instructions. If you installed from a local checkout without the system-wide step, use `./svg-auto` instead.

### Updating

Update via `go install`:

```sh
go install github.com/pedrosanto90/svg-auto@latest
```

Or pull the latest changes and rebuild from the checkout:

```sh
git pull
go build -o svg-auto .
sudo make install
```

> **Note:** `go install github.com/pedrosanto90/svg-auto@latest` requires the repository to be up to date on GitHub (commits must be pushed). As the repository has no release tags, `@latest` resolves to a pseudo-version of `main`; use `@main` to be explicit about the version.

## Supported browsers

Works with any Chromium-based browser. Automatic detection looks for, in order:

- `brave-browser`, `brave-browser-stable`
- `chromium`, `chromium-browser`
- `google-chrome`, `google-chrome-stable`
- `microsoft-edge`
- `vivaldi`, `opera`, `chrome`

## How to run

After installation, run:

```sh
svg-auto icon1.svg icon2.svg icon3.svg
```

From a local checkout (without installing):

```sh
go run . icon1.svg icon2.svg icon3.svg
```

The script imports the SVG files into IcoMoon and downloads the generated package (a `.zip` file, with the original IcoMoon name) to `./output/` (created in the current working directory). The package contains the `SVG/` and `PNG/` folders, `selection.json`, and other generated files.

To use a specific browser (path or executable name):

```sh
SVG_AUTO_BROWSER=brave-browser svg-auto icon.svg
```

To view the help:

```sh
svg-auto -h
```

