package server

import (
	"fmt"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type Symbol struct {
	Name     string
	Kind     defines.SymbolKind
	Location defines.Location

	Type *string
}

func (s Symbol) toDefines() defines.SymbolInformation {
	return defines.SymbolInformation{
		Name: s.Name,
		Kind: s.Kind,
		Location: defines.Location{
			Uri: s.Location.Uri,
			Range: defines.Range{
				Start: defines.Position{
					Line:      s.Location.Range.Start.Line,
					Character: s.Location.Range.Start.Character,
				},
				End: defines.Position{
					Line:      s.Location.Range.End.Line,
					Character: s.Location.Range.End.Character,
				},
			},
		},
	}
}

type cache struct {
	symbols map[string][]Symbol

	documents map[string]string
}

func newCache() *cache {
	return &cache{
		symbols:   make(map[string][]Symbol),
		documents: make(map[string]string),
	}
}

func (c *cache) getDocument(uri string) *string {
	content, ok := c.documents[uri]
	if !ok {
		return nil
	}
	return &content
}

func (c *cache) getSymbol(textDocumentURI string, position defines.Position) (*defines.SymbolInformation, error) {
	symbol, _ := c.findSymbolAtPosition(textDocumentURI, position)
	if symbol == nil {
		return nil, fmt.Errorf("found no symbol at position %v", position)
	}

	targetSymbol, err := c.symbolDefinition(textDocumentURI, *symbol)
	if err != nil {
		return nil, fmt.Errorf("symbol definition: %w", err)
	}

	return targetSymbol, nil
}

func (c *cache) findSymbolAtPosition(uri string, position defines.Position) (*Symbol, error) {
	symbols := c.symbols[uri]

	for _, symbol := range symbols {
		if positionBetween(position, symbol.Location.Range.Start, symbol.Location.Range.End) {
			return &symbol, nil
		}
	}
	return nil, fmt.Errorf("definition not found")
}

func (c *cache) addDocument(uri string, content string) error {
	c.documents[uri] = content

	symbols, err := parseDocument(uri, content)
	if err != nil {
		return fmt.Errorf("parse document: %w", err)
	}
	c.symbols[uri] = symbols

	return nil
}

func positionBetween(p, start, end defines.Position) bool {
	return p.Line >= start.Line &&
		p.Line <= end.Line &&
		p.Character >= start.Character &&
		p.Character <= end.Character
}

func (c *cache) symbolDefinition(uri string, symbol Symbol) (*defines.SymbolInformation, error) {
	symbols := c.symbols[uri]
	var typ string
	if symbol.Type != nil {
		typ = *symbol.Type
	}

	for _, s := range symbols {

		if s.Name == typ {
			d := s.toDefines()

			return &d, nil
		}
	}

	return nil, fmt.Errorf("symbol not found")
}
