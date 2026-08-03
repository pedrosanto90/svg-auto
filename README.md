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

## Configuration

The tool applies the downloaded icons to a project (local or remote via SSH). This requires a configuration file listing the project and the files to edit.

The config lives at `~/.config/svg-auto/config.json` (path follows [XDG](https://specifications.freedesktop.org/basedir-spec/), use the `SVG_AUTO_CONFIG` environment variable to override it). Create it with:

```sh
mkdir -p ~/.config/svg-auto && nano ~/.config/svg-auto/config.json
```

If the config is missing, the tool tells you exactly how to create it.

### Example

```json
{
  "projectPath": "user@server:/srv/app",
  "iconPrefix": "ee-icon-",
  "files": [
    {
      "name": "assets/sprite.svg",
      "mode": "text",
      "marker": "</defs>",
      "position": "before",
      "template": "<symbol id=\"{{.Prefix}}{{.Name}}\" viewBox=\"{{.ViewBox}}\">{{.Body}}</symbol>"
    },
    {
      "name": "selection.json",
      "mode": "icomoon"
    },
    {
      "name": "assets/style.css",
      "mode": "text",
      "position": "end",
      "template": "\n.{{.Prefix}}{{.Name}} {\n  width: 1.75em;\n  height: 1.75em;\n}"
    }
  ]
}
```

### Fields

- `projectPath` — path to the project. A local path (`/srv/app`) or a remote one (`user@host:/srv/app`); remote access uses your system's `ssh`. When using `~/...` on the remote host, write it as `host:~/...`.
- `iconPrefix` — prefix prepended to every icon name (e.g. `ee-icon-`). Defaults to `icon-`. It is the single source for the SVG `id`, the CSS class, and the icon tags.

### Files

Each file has a `name` (relative to `projectPath`) and a `mode`:

- `text` — renders a `template` per icon and inserts the results. Requires a `position`:
  - `end` — append to the end of the file.
  - `replace`, `before`, `after` — require a `marker`; the marker line is replaced, or the rendered block is inserted before/after the first occurrence.
  - An icon already present in the file is skipped, so re-running is safe.
- `icomoon` — edits an IcoMoon `selection.json` (project JSON) semantically: appends the new icons to `iconSets[0]`, with ids/orders continuing the existing sequence. Only the new entries are inserted; the rest of the document is left byte-for-byte untouched. Icons already present are skipped.

### Template placeholders

| Placeholder     | Description                                    |
| --------------- | ---------------------------------------------- |
| `{{.Prefix}}`   | The configured `iconPrefix`                    |
| `{{.Name}}`     | Icon name (e.g. `house-solid`)                 |
| `{{.ViewBox}}`  | The icon's `viewBox`                           |
| `{{.Body}}`     | Inner body of the `<symbol>`                   |
| `{{.Paths}}`    | Space-joined SVG path `d` values               |
| `{{.PathsJson}}`| JSON array of the path `d` values              |
| `{{.TagsJson}}` | JSON array with the icon name as a tag         |

### Behavior

- Each edited file is backed up next to itself as `<file>.orig` before any change; a backup is only made when a change is actually applied.
- After a successful run the backups are removed, so no `<file>.orig` files are left in the project repository. If the run fails partway through, the backups of the files already edited are kept so you can recover.
- Files with no new icons are left untouched.


