package server

import (
	"testing"

	"github.com/TobiasYin/go-lsp/lsp/defines"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSymbolAtPosition(t *testing.T) {
	c := newCache()
	c.symbols["myfile.proto"] = []Symbol{
		{
			Name: "MyMessage",
			Kind: defines.SymbolKindStruct,
			Location: defines.Location{
				Uri: "myfile.proto",
				Range: defines.Range{
					Start: defines.Position{
						Line:      0,
						Character: 0,
					},
					End: defines.Position{
						Line:      0,
						Character: 10,
					},
				},
			},
		},
	}

	symbol, err := c.findSymbolAtPosition("myfile.proto", defines.Position{Line: 0, Character: 5})
	require.NoError(t, err)
	require.NotNil(t, symbol)
	assert.Equal(t, "MyMessage", symbol.Name)
}
