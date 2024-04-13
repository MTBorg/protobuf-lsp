package server

import (
	"context"
	"testing"

	"github.com/TobiasYin/go-lsp/lsp/defines"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefinition(t *testing.T) {
	ctx := context.Background()
	symbols := []Symbol{
		&MessageSymbol{
			name: "Foo",
			SymbolBase: SymbolBase{
				Uri: "file:///test.proto",
				Loc: defines.Location{
					Range: defines.Range{
						Start: defines.Position{
							Line:      0,
							Character: 0,
						},
						End: defines.Position{
							Line:      10,
							Character: 0,
						},
					},
				},
			},
		},
		&MessageSymbol{
			name: "Bar",
			SymbolBase: SymbolBase{
				Uri: "file:///test.proto",
				Loc: defines.Location{
					Range: defines.Range{
						Start: defines.Position{
							Line:      20,
							Character: 0,
						},
						End: defines.Position{
							Line:      30,
							Character: 0,
						},
					},
				},
			},
		},
		&FieldSymbol{
			Name:     "foo",
			TypeName: "Foo",
			SymbolBase: SymbolBase{
				Uri: "file:///test.proto",
				Loc: defines.Location{
					Range: defines.Range{
						Start: defines.Position{
							Line:      21,
							Character: 0,
						},
						End: defines.Position{
							Line:      21,
							Character: 20,
						},
					},
				},
			},
		},
	}
	s := &server{
		symbols: symbols,
	}

	locationLinks, err := s.Definition(ctx, &defines.DefinitionParams{
		TextDocumentPositionParams: defines.TextDocumentPositionParams{
			TextDocument: defines.TextDocumentIdentifier{
				Uri: "file:///test.proto",
			},
			Position: defines.Position{
				Line:      21,
				Character: 10,
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, locationLinks)
	require.Len(t, *locationLinks, 1)
	link := (*locationLinks)[0]
	assert.EqualValues(t, "file:///test.proto", link.TargetUri)
	assert.Equal(t, defines.Range{
		Start: defines.Position{
			Line:      0,
			Character: 0,
		},
		End: defines.Position{
			Line:      10,
			Character: 0,
		},
	}, link.TargetRange)
}
