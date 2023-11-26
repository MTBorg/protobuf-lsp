# Protobuf-lsp

Protobuf language server written in Go. Mostly meant as a toy project so don't
expect any stability.

## Implemented features

- [Go to definition](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_definition)(WIP)
- [Find references](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_references)(WIP)

## Usage

### Coc.nvim

#### http-mode

This requires the language server to be started before connecting with the
client:

```shell
go run cmd/server/main.go --mode=http
```

Add this `lanugageserver`-configuration to your `coc-settings.json`-file:

```json
"languageserver": [
    "proto-http": {
      "filetypes": ["proto"],
      "host": "127.0.0.1",
      "port": 8080
    }
]
```

#### stdio-mode

This requires a binary of the language server.
Add this `lanugageserver`-configuration to your `coc-settings.json`-file:

```json
"languageserver": [
    "proto-stdio": {
      "filetypes": ["proto"],
      "command": "<path to your binary>",
    }
]
```
