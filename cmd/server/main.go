package main

import (
	"flag"
	"log"

	protolsp "github.com/MTBorg/protobuf-lsp/protolsp"
	protoserver "github.com/MTBorg/protobuf-lsp/protolsp/server"
	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

var address = flag.String("address", "localhost:8080", "address to listen on")

func main() {
	flag.Parse()
	logs.Init(log.Default())

	server := lsp.NewServer(&lsp.Options{
		Network: "tcp",
		Address: *address,
		CompletionProvider: &defines.CompletionOptions{
			TriggerCharacters: &[]string{"."},
		}})

	protobuf := protoserver.NewServer()
	withProtobuf(server, protobuf)

	server.Run()
}

func withProtobuf(server *lsp.Server, protobuf protolsp.ProtoLSP) {
	server.OnDefinition(protobuf.Definition)
	server.OnDidChangeTextDocument(protobuf.TextDocumentDidChange)
	server.OnDidOpenTextDocument(protobuf.TextDocumentDidOpen)
}
