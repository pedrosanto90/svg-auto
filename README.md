# svg-auto

Automation for [IcoMoon](https://icomoon.io) to import SVG files and download the generated package, using [chromedp](https://github.com/chromedp/chromedp).

## Supported browsers

Works with any Chromium-based browser. Automatic detection looks for, in order:

- `brave-browser`, `brave-browser-stable`
- `chromium`, `chromium-browser`
- `google-chrome`, `google-chrome-stable`
- `microsoft-edge`
- `vivaldi`, `opera`, `chrome`

## How to run

```sh
go run . icon1.svg icon2.svg icon3.svg
```

The script imports the SVG files into IcoMoon and downloads the generated package (a `.zip` file, with the original IcoMoon name) to `./output/`. The package contains the `SVG/` and `PNG/` folders, `selection.json`, and other generated files.

To use a specific browser (path or executable name):

```sh
SVG_AUTO_BROWSER=brave-browser go run . icon.svg
```

To view the help:

```sh
go run . -h
```

## Requirements

- Go 1.26+
- A Chromium-based browser installed (or set via `SVG_AUTO_BROWSER`)
