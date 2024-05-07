package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type server struct {
	documents map[string]string
	symbols   SymbolLookup
}

func NewServer() *server {
	return &server{
		documents: make(map[string]string),
	}
}

func (s *server) Definition(ctx context.Context, params *defines.DefinitionParams) (*[]defines.LocationLink, error) {
	uri := string(params.TextDocument.Uri)
	position := params.Position
	symbol, err := s.getSymbol(uri, position)
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

	document, ok := s.documents[uri]
	if !ok {
		return fmt.Errorf("document not found")
	}

	newDocument := document
	for _, change := range params.ContentChanges {
		d, err := applyChange(newDocument, change)
		if err != nil {
			return err
		}
		newDocument = d
	}

	if err := s.addDocument(uri, newDocument); err != nil {
		return fmt.Errorf("update cache: %w", err)
	}

	return nil
}

func (s *server) TextDocumentDidOpen(ctx context.Context, params *defines.DidOpenTextDocumentParams) error {
	uri := string(params.TextDocument.Uri)
	if err := s.addDocument(uri, params.TextDocument.Text); err != nil {
		return fmt.Errorf("add document: %w", err)
	}
	return nil
}

func (s *server) References(ctx context.Context, params *defines.ReferenceParams) (*[]defines.Location, error) {
	uri := string(params.TextDocument.Uri)
	position := params.Position

	locations, err := s.getSymbolReferences(uri, position)
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

func (s *server) addDocument(uri string, content string) error {
	s.documents[uri] = content

	symbols, err := parseDocument(uri, content)
	if err != nil {
		return fmt.Errorf("parse document: %w", err)
	}

	s.symbols = s.symbols.mergeWithURI(symbols, uri)

	importSymbols := SymbolLookup(symbols).BySymbolType(&ImportSymbol{})

	for _, symbol := range importSymbols {
		importSymbol := symbol.(*ImportSymbol)
		content, err := os.ReadFile("./" + importSymbol.ImportPath)
		if err != nil {
			return fmt.Errorf("read import: %w", err)
		}

		if err := s.addDocument(uriFromImportPath(importSymbol.ImportPath), string(content)); err != nil {
			return fmt.Errorf("add import: %w", err)
		}
	}

	logger.Debug("cache refreshed", slog.String("uri", uri))

	return nil
}

func (s *server) getSymbol(textDocumentURI string, position defines.Position) (*defines.SymbolInformation, error) {
	cursorSymbol := symbolUnderCursor(s.symbols, position)
	if cursorSymbol == nil {
		return nil, fmt.Errorf("found no symbol under cursor at position %+v", position)
	}

	targetSymbol := *cursorSymbol

	if typeableSymbol, ok := (*cursorSymbol).(Typeable); ok {
		s, err := s.typeDefinition(typeableSymbol)
		if err != nil {
			return nil, fmt.Errorf("symbol definition: %w", err)
		}
		targetSymbol = s
	}

	d := toDefines(targetSymbol)
	return &d, nil
}

func (s *server) getSymbolReferences(textDocumentURI string, position defines.Position) (*[]defines.Location, error) {
	cursorSymbol := symbolUnderCursor(s.symbols, position)
	if cursorSymbol == nil {
		return nil, fmt.Errorf("found no symbol under cursor at position %+v", position)
	}

	rootSymbol := *cursorSymbol
	if typeableSymbol, ok := (*cursorSymbol).(Typeable); ok {
		s, err := s.typeDefinition(typeableSymbol)
		if err != nil {
			return nil, fmt.Errorf("symbol definition: %w", err)
		}
		rootSymbol = s
	}

	if _, ok := rootSymbol.(Nameable); !ok {
		return nil, fmt.Errorf("symbol %T is not nameable", rootSymbol)
	}
	name := rootSymbol.(Nameable).Name()

	symbols := FilterUsingSymbolTypePredicate(s.symbols, func(m Typeable) bool {
		return m.Type() == name
	})

	var result []defines.Location
	for _, symbol := range symbols {
		result = append(result, symbol.Location())
	}

	return &result, nil
}

func (s *server) typeDefinition(symbol Typeable) (*MessageSymbol, error) {
	match := FilterUsingSymbolTypePredicate(s.symbols, func(m *MessageSymbol) bool {
		return m.Name() == symbol.Type()
	}).First()
	if match == nil {
		return nil, fmt.Errorf("symbol %q not found", symbol.Type())
	}
	return (*match).(*MessageSymbol), nil
}

func positionBetween(p, start, end defines.Position) bool {
	switch {
	case p.Line > start.Line && p.Line < end.Line:
		return true
	case p.Line == start.Line && p.Line == end.Line:
		return p.Character >= start.Character && p.Character <= end.Character
	case p.Line == start.Line:
		return p.Character >= start.Character
	case p.Line == end.Line:
		return p.Character <= end.Character
	default:
		return false
	}
}

func uriFromImportPath(importPath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("file://%s/%s", cwd, importPath)
}

func symbolUnderCursor(symbols []Symbol, cursorPosition defines.Position) *Symbol {
	var potentials []Symbol
	for _, symbol := range symbols {
		if positionBetween(cursorPosition, symbol.Location().Range.Start, symbol.Location().Range.End) {
			potentials = append(potentials, symbol)
		}
	}

	if len(potentials) == 0 {
		return nil
	}

	// Find the symbol closest to the cursor
	closest := 0
	for i, symbol := range potentials[1:] {
		if symbol.Location().Range.Start.Line > potentials[closest].Location().Range.Start.Line {
			closest = i + 1
		}
	}

	return &potentials[closest]
}
