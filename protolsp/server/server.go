package server

import (
	"context"
	"fmt"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type server struct {
	cache *cache
}

func NewServer() *server {
	return &server{
		cache: newCache(),
	}
}

func (s *server) Definition(ctx context.Context, params *defines.DefinitionParams) (*[]defines.LocationLink, error) {
	uri := string(params.TextDocument.Uri)
	position := params.Position
	symbol, err := s.cache.getSymbol(uri, position)
	if err != nil {
		return nil, err
	}

	link := defines.LocationLink{
		TargetUri:            symbol.Location.Uri,
		TargetRange:          symbol.Location.Range,
		TargetSelectionRange: symbol.Location.Range,
	}

	result := &[]defines.LocationLink{link}
	return result, nil
}

func (s *server) TextDocumentDidChange(ctx context.Context, params *defines.DidChangeTextDocumentParams) error {
	uri := string(params.TextDocument.Uri)

	document := s.cache.getDocument(uri)
	if document == nil {
		return fmt.Errorf("document not found")
	}

	newDocument := *document
	for _, change := range params.ContentChanges {
		d, err := applyChange(newDocument, change)
		if err != nil {
			return err
		}
		newDocument = d
	}

	if err := s.cache.addDocument(uri, newDocument); err != nil {
		return fmt.Errorf("update cache: %w", err)
	}

	return nil
}

func (s *server) TextDocumentDidOpen(ctx context.Context, params *defines.DidOpenTextDocumentParams) error {
	uri := string(params.TextDocument.Uri)
	s.cache.addDocument(uri, params.TextDocument.Text)
	return nil
}

func applyChange(content string, change defines.TextDocumentContentChangeEvent) (string, error) {
	if s, ok := change.Text.(string); ok {
		return s, nil
	}

	return "", fmt.Errorf("diff changes not implemented")
}
