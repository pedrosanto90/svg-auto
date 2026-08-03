# svg-auto

Automation for [IcoMoon](https://icomoon.io) that imports SVG files, downloads the generated package, and applies the resulting icons to a project — local or remote over SSH. It drives a real Chromium-based browser with [chromedp](https://github.com/chromedp/chromedp).

## Features

- Imports any number of SVG files into IcoMoon automatically.
- Downloads the generated package (`.zip`) with the original IcoMoon name.
- Extracts the icons (symbols and paths) from the package.
- Applies the new icons to a project by editing its SVG sprite, IcoMoon `selection.json`, and CSS — configured via a single JSON file.
- Supports local projects and remote projects over SSH (uses your system's `ssh`).
- Idempotent: icons already present are skipped, so re-running is safe.
- Backs up every edited file (`.orig`) and cleans up the backups after a successful run.

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

The Makefile installs to `/usr/local/bin` by default; override with `make install PREFIX=...`.

### Option 2: go install

Install directly with the Go toolchain. Requires Go 1.26+:

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

## Usage

### Basic usage

```sh
svg-auto icon1.svg icon2.svg icon3.svg
```

From a local checkout (without installing):

```sh
go run . icon1.svg icon2.svg icon3.svg
```

### Options

```
Usage: svg-auto <file1.svg> [file2.svg ...]

Imports SVG files into IcoMoon and downloads the generated package (.zip) to ./output/.

Options:
  -h, --help    show this help

Environment variables:
  SVG_AUTO_BROWSER    path or name of the browser executable (optional)
```

### Environment variables

| Variable           | Description                                                             |
| ------------------ | ----------------------------------------------------------------------- |
| `SVG_AUTO_BROWSER` | Path or name of the browser executable, to override automatic detection |
| `SVG_AUTO_CONFIG`  | Path to the config file, to override the default location               |

### How it works

1. The tool opens IcoMoon in a browser and imports each SVG file.
2. It downloads the generated package (a `.zip` file with the original IcoMoon name) into `./output/download/`.
3. It extracts the icons (symbols and SVG paths) from the package.
4. It loads the config and applies the icons to your project, editing the configured files (see [Configuration](#configuration)).
5. The zip is moved to `./output/` and the temporary download directory is removed.

The output directory contains the downloaded package — with the `SVG/` and `PNG/` folders, `selection.json`, and the other generated files.

To use a specific browser (path or executable name):

```sh
SVG_AUTO_BROWSER=brave-browser svg-auto icon.svg
```

### Supported browsers

Works with any Chromium-based browser. Automatic detection looks for, in order:

- `brave-browser`, `brave-browser-stable`
- `chromium`, `chromium-browser`
- `google-chrome`, `google-chrome-stable`
- `microsoft-edge`
- `vivaldi`, `opera`, `chrome`

It also checks common install locations on Linux (`/usr/bin`, `/snap/bin`), macOS (`/Applications`), and Windows (`Program Files`). If no browser is found, set `SVG_AUTO_BROWSER` to the executable path.

## Configuration

The tool applies the downloaded icons to a project (local or remote via SSH). This requires a configuration file listing the project and the files to edit.

### Config location

The config lives at `~/.config/svg-auto/config.json` (path follows [XDG](https://specifications.freedesktop.org/basedir-spec/), use the `SVG_AUTO_CONFIG` environment variable to override it). Create it with:

```sh
mkdir -p ~/.config/svg-auto && nano ~/.config/svg-auto/config.json
```

If the config is missing or invalid, the tool prints an error explaining exactly what to create or fix.

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
      "separator": true,
      "template": "\n.{{.Prefix}}{{.Name}} {\n  width: 1.75em;\n  height: 1.75em;\n}"
    }
  ]
}
```

### Fields

| Field          | Description                                                                                                                                                                                         |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `projectPath`  | Path to the project. A local path (`/srv/app`) or a remote one (`user@host:/srv/app`); remote access uses your system's `ssh`. When using `~/...` on the remote host, write it as `host:~/...`. Required. |
| `iconPrefix`   | Prefix prepended to every icon name (e.g. `ee-icon-`). Defaults to `icon-`. It is the single source for the SVG `id`, the CSS class, and the icon tags.                                              |
| `files`        | List of files to edit, relative to `projectPath`. At least one file is required.                                                                                                                    |

### File rules

Each entry in `files` has a `name` (relative to `projectPath`) and a `mode`:

| Field      | Description                                                                                                                    |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `name`     | Path of the file, relative to `projectPath`. Required.                                                                          |
| `mode`     | `text` or `icomoon`. Required.                                                                                                  |
| `marker`   | Text used as an insertion point (`text` mode). Required for `before`, `after`, and `replace`.                                   |
| `position` | `end`, `before`, `after`, or `replace` (`text` mode). Required for `text`.                                                      |
| `template` | Go template rendered once per icon (`text` mode). Required.                                                                     |
| `separator`| When `true`, inserts a blank line between the existing content and the first new entry (`text` mode, optional).                  |

#### text mode

Renders the `template` per icon and inserts the results, depending on `position`:

- `end` — appends to the end of the file.
- `replace` — replaces the first occurrence of `marker` with the rendered block.
- `before` / `after` — inserts the rendered block before/after the first occurrence of `marker`.

An icon already present in the file is skipped, so re-running is safe.

`separator: true` inserts a blank line between the existing content and the first new entry — useful for files such as CSS, where each rule should be separated by a blank line.

#### icomoon mode

Edits an IcoMoon `selection.json` (project JSON) semantically: appends the new icons to `iconSets[0]`, with ids and orders continuing the existing sequence. Only the new entries are inserted; the rest of the document is left byte-for-byte untouched. Icons already present are skipped.

### Template placeholders

| Placeholder      | Description                              |
| ---------------- | ---------------------------------------- |
| `{{.Prefix}}`    | The configured `iconPrefix`              |
| `{{.Name}}`      | Icon name (e.g. `house-solid`)           |
| `{{.ViewBox}}`   | The icon's `viewBox`                     |
| `{{.Body}}`      | Inner body of the `<symbol>`             |
| `{{.Paths}}`     | Comma-joined SVG path `d` values         |
| `{{.PathsJson}}` | JSON array of the path `d` values        |
| `{{.TagsJson}}`  | JSON array with the icon name as a tag   |

### SSH and remote projects

To apply icons to a project on a remote machine, set `projectPath` to `user@host:/path/to/project`. The tool runs `ssh` (your system's SSH client, using your existing keys and `~/.ssh/config`) for every read and write. To reference the remote home directory, write `host:~/path`.

### Behavior

- Each edited file is backed up next to itself as `<file>.orig` before any change; a backup is only made when a change is actually applied.
- After a successful run the backups are removed, so no `<file>.orig` files are left in the project repository. If the run fails partway through, the backups of the files already edited are kept so you can recover.
- Files with no new icons are left untouched.

## Troubleshooting

| Problem                          | Solution                                                       |
| -------------------------------- | -------------------------------------------------------------- |
| `no supported browser was found` | Install a Chromium-based browser, or set `SVG_AUTO_BROWSER` to the executable path. |
| `config not found at ...`        | Create the config file as instructed by the error message (see [Config location](#config-location)). |
| `invalid config ...`             | Check the error message; it points to the missing or wrong field. |

## License

[MIT](LICENSE)
