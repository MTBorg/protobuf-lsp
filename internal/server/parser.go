package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/TobiasYin/go-lsp/lsp/defines"
	protoparser "github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/parser"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"
)

type handler func(T parser.Visitee) []Symbol

func parseDocument(uri string, content string) ([]Symbol, error) {
	r := strings.NewReader(content)
	got, err := protoparser.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse, err %v\n", err)
	}

	symbols := Walk(got,
		withHandler(handleMessage),
		withHandler(handleField),
		withHandler(handleImport),
		withHandler(handleRPC),
	)

	for _, symbol := range symbols {
		symbol.SetURI(uri)
	}

	// prettyPrint(symbols)

	return symbols, nil
}

func prettyPrint(symbols []Symbol) {
	d, err := json.MarshalIndent(symbols, "", "  ")
	if err != nil {
		panic(err)
	}
	logger.Debug("symbols", slog.String("symbols", string(d)))
}

func Walk(proto *parser.Proto, handlers ...handler) []Symbol {
	return walk(proto.ProtoBody, handlers...)
}

func walk(body []parser.Visitee, handlers ...handler) []Symbol {
	var symbols []Symbol
	for _, element := range body {
		for _, handler := range handlers {
			s := handler(element)
			symbols = append(symbols, s...)
		}

		chilren := getChildren(element)
		s := walk(chilren, handlers...)
		symbols = append(symbols, s...)
	}
	return symbols
}

func withHandler[V parser.Visitee](apply func(v V) []Symbol) handler {
	return func(v parser.Visitee) []Symbol {
		if s, ok := v.(V); ok {
			return apply(s)
		}
		return nil
	}
}

func getChildren(v parser.Visitee) []parser.Visitee {
	switch t := v.(type) {
	case *parser.Message:
		return t.MessageBody
	case *parser.Enum:
		return t.EnumBody
	case *parser.Service:
		return t.ServiceBody
	}
	return nil
}

func symbolFromMessage(m *parser.Message) *MessageSymbol {
	return &MessageSymbol{
		SymbolBase: newSymbolBase(m.Meta),
		name:       m.MessageName,
	}
}

func symbolFromField(f *parser.Field) *FieldSymbol {
	return &FieldSymbol{
		SymbolBase: newSymbolBase(f.Meta),
		Name:       f.FieldName,
		TypeName:   f.Type,
	}
}

func symbolFromImport(i *parser.Import) *ImportSymbol {
	return &ImportSymbol{
		SymbolBase: newSymbolBase(i.Meta),
		ImportPath: strings.Trim(i.Location, "\""),
	}
}

func symbolFromRPCRequest(r *parser.RPCRequest) *RPCRequestSymbol {
	return &RPCRequestSymbol{
		TypeName:   r.MessageType,
		SymbolBase: newSymbolBase(r.Meta),
	}
}

func symbolFromRPCResponse(r *parser.RPCResponse) *RPCResponseSymbol {
	return &RPCResponseSymbol{
		TypeName:   r.MessageType,
		SymbolBase: newSymbolBase(r.Meta),
	}
}

func parserPositionToDefinesPosition(p meta.Position) defines.Position {
	return defines.Position{
		Line:      uint(p.Line) - 1,
		Character: uint(p.Column) - 1,
	}
}

func newSymbolBase(meta meta.Meta) SymbolBase {
	return SymbolBase{
		Loc: defines.Location{
			Range: defines.Range{
				Start: parserPositionToDefinesPosition(meta.Pos),
				End:   parserPositionToDefinesPosition(meta.LastPos),
			},
		},
	}
}

func handleMessage(m *parser.Message) []Symbol {
	return []Symbol{symbolFromMessage(m)}
}

func handleField(f *parser.Field) []Symbol {
	return []Symbol{symbolFromField(f)}
}

func handleImport(i *parser.Import) []Symbol {
	return []Symbol{symbolFromImport(i)}
}

func handleRPC(r *parser.RPC) []Symbol {
	var symbols []Symbol
	if r.RPCRequest != nil {
		symbols = append(symbols, symbolFromRPCRequest(r.RPCRequest))
	}

	if r.RPCResponse != nil {
		symbols = append(symbols, symbolFromRPCResponse(r.RPCResponse))
	}

	return symbols
}
