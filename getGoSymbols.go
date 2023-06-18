package main

import (
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"go/ast"
	"go/parser"
	"go/token"
)



// Request format for GetGoSymbols
type GetGoSymbolsRequest struct {
	FilePath string `json:"file_path,omitempty"`
}

// Function that gets Go symbols in a file
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
		if n != node && ast.Node(n).Parent() != node { return true }
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
		}
		return true
	})

	return map[string][]string{
		"symbols": symbols,
	}
}

// Metadata for GetGoSymbols
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