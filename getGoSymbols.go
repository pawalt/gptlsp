package main

import (
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"go/ast"
	"go/parser"
	"go/token"
)

// GetGoSymbolsRequest defines the request format for GetGoSymbols function
type GetGoSymbolsRequest struct {
	FilePath string `json:"file_path,omitempty"`
}

// GetGoSymbols gets Go symbols in a file
func GetGoSymbols(raw string) map[string][]string {
	var req GetGoSymbolsRequest
	_ = json.Unmarshal([]byte(raw), &req)
	fp := req.FilePath

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fp, nil, parser.ParseComments)
	if err != nil {
		return map[string][]string{
			"error": {err.Error()},
		}
	}

	var symbols []string
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.VAR {
				for _, spec := range x.Specs {
					vspec := spec.(*ast.ValueSpec)
					symbols = append(symbols, vspec.Names[0].Name) 
				}
			}
		case *ast.FuncDecl:
			symbols = append(symbols, x.Name.Name)
			return false
		}
		return true
	})

	return map[string][]string{
		"symbols": symbols,
	}
}

// getGoSymbolsMetadata defines metadata for GetGoSymbols function
var getGoSymbolsMetadata = &openai.FunctionDefine{
	Name:        "get_go_symbols",
	Description: "Get Go symbols in a file",
	Parameters: &openai.FunctionParams{
		Type: openai.JSONSchemaTypeObject,
		Properties: map[string]*openai.JSONSchemaDefine{
			"file_path": {
				Type:        openai.JSONSchemaTypeString,
				Description: `Relative path to the Go file`,
			},
		},
		Required: []string{"file_path"},
	},
}