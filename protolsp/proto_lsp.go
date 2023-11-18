package protolsp

import (
	"context"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type ProtoLSP interface {
	TextDocumentDidOpen(context.Context, *defines.DidOpenTextDocumentParams) error
	TextDocumentDidChange(context.Context, *defines.DidChangeTextDocumentParams) error
	Definition(context.Context, *defines.DefinitionParams) (*[]defines.LocationLink, error)
}
