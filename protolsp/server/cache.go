package server

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type cache struct {
	documents map[string]string

	lookup SymbolLookup
}

func newCache() *cache {
	return &cache{
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
	cursorSymbol := symbolUnderCursor(c.lookup, position)
	if cursorSymbol == nil {
		return nil, fmt.Errorf("found no symbol under cursor at position %+v", position)
	}

	targetSymbol, err := c.typeDefinition(*cursorSymbol)
	if err != nil {
		return nil, fmt.Errorf("symbol definition: %w", err)
	}

	d := toDefines(targetSymbol)
	return &d, nil
}

func (c *cache) getSymbolReferences(textDocumentURI string, position defines.Position) (*[]defines.Location, error) {
	cursorSymbol := symbolUnderCursor(c.lookup, position)
	if cursorSymbol == nil {
		return nil, fmt.Errorf("found no symbol under cursor at position %+v", position)
	}

	rootSymbol := *cursorSymbol
	if _, ok := (*cursorSymbol).(MessageSymbol); !ok {
		s, err := c.typeDefinition(*cursorSymbol)
		if err != nil {
			return nil, fmt.Errorf("symbol definition: %w", err)
		}
		rootSymbol = s
	}

	if _, ok := rootSymbol.(Nameable); !ok {
		return nil, fmt.Errorf("symbol %T is not nameable", rootSymbol)
	}
	name := rootSymbol.(Nameable).Name()

	symbols := FilterUsingSymbolTypePredicate(c.lookup, func(m FieldSymbol) bool {
		return m.Type == name
	})

	var result []defines.Location
	for _, symbol := range symbols {
		result = append(result, symbol.Location())
	}

	return &result, nil
}

func (c *cache) addDocument(uri string, content string) error {
	c.documents[uri] = content

	symbols, err := parseDocument(uri, content)
	if err != nil {
		return fmt.Errorf("parse document: %w", err)
	}

	c.lookup = c.mergeSymbolsFromFile(uri, symbols)

	importSymbols := SymbolLookup(symbols).BySymbolType(ImportSymbol{})

	for _, symbol := range importSymbols {
		importSymbol := symbol.(ImportSymbol)
		content, err := os.ReadFile("./" + importSymbol.ImportPath)
		if err != nil {
			return fmt.Errorf("read import: %w", err)
		}

		if err := c.addDocument(uriFromImportPath(importSymbol.ImportPath), string(content)); err != nil {
			return fmt.Errorf("add import: %w", err)
		}
	}

	logger.Debug("cache refreshed", slog.String("uri", uri))

	return nil
}

func (c *cache) mergeSymbolsFromFile(fileURI string, symbols []Symbol) []Symbol {
	var result []Symbol

	// Remove all symbols associated with the fileURI
	for _, symbol := range c.lookup {
		if string(symbol.Location().Uri) != fileURI {
			result = append(result, symbol)
		}
	}

	// Add all symbols from the fileURI
	for _, symbol := range symbols {
		result = append(result, symbol)
	}

	return result
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

func (c *cache) typeDefinition(symbol Symbol) (Symbol, error) {
	fieldSymbol, ok := symbol.(FieldSymbol)
	if !ok {
		return nil, fmt.Errorf("don't know how to lookup definition for symbol %T", symbol)
	}

	match := FilterUsingSymbolTypePredicate(c.lookup, func(m MessageSymbol) bool {
		return m.Name() == fieldSymbol.Type
	}).First()
	if match == nil {
		return nil, fmt.Errorf("symbol %q not found", fieldSymbol.Type)
	}
	return *match, nil
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
