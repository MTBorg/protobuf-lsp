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

type handler func(v parser.Visitee)

func parseDocument(uri string, content string) ([]Symbol, error) {
	r := strings.NewReader(content)
	got, err := protoparser.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse, err %v\n", err)
	}

	var symbols []Symbol
	Walk(got,
		withHandler(func(m *parser.Message) {
			symbol := symbolFromMessage(m)
			symbol.Uri = uri
			symbols = append(symbols, symbol)
		}),
		withHandler(func(v *parser.Field) {
			symbol := symbolFromField(v)
			symbol.Uri = uri
			symbols = append(symbols, symbol)
		}),
		withHandler(func(v *parser.Import) {
			symbol := symbolFromImport(v)
			symbol.Uri = uri
			symbols = append(symbols, symbol)
		}),
		withHandler(func(v *parser.RPC) {
			if v.RPCRequest != nil {
				requestSymbol := symbolFromRPCRequest(v.RPCRequest)
				requestSymbol.Uri = uri
				symbols = append(symbols, requestSymbol)
			}

			if v.RPCResponse != nil {
				responseSymbol := symbolFromRPCResponse(v.RPCResponse)
				responseSymbol.Uri = uri
				symbols = append(symbols, responseSymbol)
			}
		}),
	)

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

func Walk(proto *parser.Proto, handlers ...handler) {
	walk(proto.ProtoBody, handlers...)
}

func walk(body []parser.Visitee, handlers ...handler) {
	for _, element := range body {
		for _, handler := range handlers {
			handler(element)
		}

		chilren := getChildren(element)
		walk(chilren, handlers...)
	}
}

func withHandler[V parser.Visitee](apply func(v V)) handler {
	return func(v parser.Visitee) {
		if s, ok := v.(V); ok {
			apply(s)
		}
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

func symbolFromMessage(m *parser.Message) MessageSymbol {
	return MessageSymbol{
		SymbolBase: newSymbolBase(m.Meta),
		name:       m.MessageName,
	}
}

func symbolFromField(f *parser.Field) FieldSymbol {
	return FieldSymbol{
		SymbolBase: newSymbolBase(f.Meta),
		Name:       f.FieldName,
		TypeName:   f.Type,
	}
}

func symbolFromImport(i *parser.Import) ImportSymbol {
	return ImportSymbol{
		SymbolBase: newSymbolBase(i.Meta),
		ImportPath: strings.Trim(i.Location, "\""),
	}
}

func symbolFromRPCRequest(r *parser.RPCRequest) RPCRequestSymbol {
	return RPCRequestSymbol{
		TypeName:   r.MessageType,
		SymbolBase: newSymbolBase(r.Meta),
	}
}

func symbolFromRPCResponse(r *parser.RPCResponse) RPCResponseSymbol {
	return RPCResponseSymbol{
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
