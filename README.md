# Protobuf-lsp

Protobuf language server written in Go. Mostly meant as a toy project so don't
expect any stability.

## Implemented features

- [Go to definition](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_definition)(WIP)

## Usage

At the moment the server is run over tcp as a standalone process. Start the
server by running:

```shell
go run cmd/server/main.go --address=<address>
```

### Coc.nvim

Add this `lanugageserver`-configuration to your `coc-settings.json`-file:

```json
"languageserver": [
    "proto": {
      "filetypes": ["proto"],
      "host": "127.0.0.1",
      "port": 8080
    }
]
```
