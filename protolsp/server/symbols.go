package server

import (
	"reflect"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type Symbol interface {
	Location() defines.Location
	URI() string
}

type Nameable interface {
	Name() string
}

type SymbolBase struct {
	// Don't use the URI from the Location, because it seems to always be empty
	// TODO: Investigate why
	Loc defines.Location
	Uri string
}

func (b SymbolBase) Location() defines.Location {
	uri := b.Uri
	loc := b.Loc
	loc.Uri = defines.DocumentUri(uri)
	return loc
}

func (b SymbolBase) URI() string {
	return string(b.Uri)
}

type SymbolLookup []Symbol

func (l SymbolLookup) ByURI(uri string) SymbolLookup {
	var result []Symbol
	for _, symbol := range l {
		if symbol.URI() == uri {
			result = append(result, symbol)
		}
	}
	return result
}

func (l SymbolLookup) BySymbolType(typ Symbol) SymbolLookup {
	var result SymbolLookup
	typType := reflect.TypeOf(typ)

	// Iterate through symbols and filter by type
	for _, symbol := range l {
		symbolType := reflect.TypeOf(symbol)
		if symbolType == typType {
			result = append(result, symbol)
		}
	}

	return result
}

func (l SymbolLookup) AtPosition(p defines.Position) SymbolLookup {
	var result SymbolLookup
	for _, symbol := range l {
		if positionBetween(p, symbol.Location().Range.Start, symbol.Location().Range.End) {
			result = append(result, symbol)
		}
	}
	return result
}

func FilterUsingSymbolTypePredicate[T Symbol](l SymbolLookup, pred func(t T) bool) SymbolLookup {
	var result SymbolLookup
	for _, symbol := range l {
		if t, ok := symbol.(T); ok && pred(t) {
			result = append(result, symbol)
		}
	}
	return result
}

func (l SymbolLookup) First() *Symbol {
	if len(l) == 0 {
		return nil
	}
	return &l[0]
}

type FieldSymbol struct {
	SymbolBase

	Name string
	Type string
}

type ImportSymbol struct {
	SymbolBase

	ImportPath string
}

type MessageSymbol struct {
	SymbolBase

	name string
}

func (m MessageSymbol) Name() string {
	return m.name
}

func toDefines(symbol Symbol) defines.SymbolInformation {
	info := defines.SymbolInformation{
		Location: defines.Location{
			Uri:   defines.DocumentUri(symbol.URI()),
			Range: symbol.Location().Range,
		},
	}

	switch t := symbol.(type) {
	case FieldSymbol:
		info.Name = t.Name
		info.Kind = defines.SymbolKindField
	}

	return info
}

func (l SymbolLookup) mergeWithURI(symbols []Symbol, uri string) SymbolLookup {
	var result []Symbol

	// Remove all symbols associated with the fileURI
	for _, symbol := range l {
		if string(symbol.Location().Uri) != uri {
			result = append(result, symbol)
		}
	}

	// Add all symbols from the fileURI
	for _, symbol := range symbols {
		result = append(result, symbol)
	}

	return result
}
