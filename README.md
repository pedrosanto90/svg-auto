# svg-auto

Automação do [IcoMoon](https://icomoon.io) para importar ficheiros SVG e descarregar o pacote gerado, usando [chromedp](https://github.com/chromedp/chromedp).

## Browsers suportados

Funciona com qualquer browser baseado em Chromium. A deteção automática procura, por ordem:

- `brave-browser`, `brave-browser-stable`
- `chromium`, `chromium-browser`
- `google-chrome`, `google-chrome-stable`
- `microsoft-edge`
- `vivaldi`, `opera`, `chrome`

## Como correr

```sh
go run . icon1.svg icon2.svg icon3.svg
```

O script importa os ficheiros SVG para o IcoMoon e descarrega o pacote gerado (ficheiro `.zip`, com o nome original do IcoMoon) para `./output/`. O pacote contém as pastas `SVG/`, `PNG/`, `selection.json` e outros ficheiros gerados.

Para usar um browser específico (caminho ou nome do executável):

```sh
SVG_AUTO_BROWSER=brave-browser go run . icon.svg
```

Para ver a ajuda:

```sh
go run . -h
```

## Requisitos

- Go 1.26+
- Um browser Chromium-based instalado (ou definido via `SVG_AUTO_BROWSER`)
