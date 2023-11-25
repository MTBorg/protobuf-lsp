package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type server struct {
	cache *cache
}

var logger slog.Logger

func NewServer() *server {
	return &server{
		cache: newCache(),
	}
}

func InitLogger(l slog.Logger) {
	logger = l
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
	if err := s.cache.addDocument(uri, params.TextDocument.Text); err != nil {
		return fmt.Errorf("add document: %w", err)
	}
	return nil
}

func (s *server) References(ctx context.Context, params *defines.ReferenceParams) (*[]defines.Location, error) {
	uri := string(params.TextDocument.Uri)
	position := params.Position

	locations, err := s.cache.getSymbolReferences(uri, position)
	if err != nil {
		return nil, err
	}
	return locations, nil
}

func applyChange(content string, change defines.TextDocumentContentChangeEvent) (string, error) {
	if s, ok := change.Text.(string); ok {
		return s, nil
	}

	return "", fmt.Errorf("diff changes not implemented")
}
