package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

func GetGoSymbols(filepath string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var symbols []string
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if x.Obj != nil && x.Obj.Kind == ast.Var {
				symbols = append(symbols, x.Name)
			}
		}
		return true
	})

	return symbols, nil
}

func main() {
	filepath := os.Args[1]
	symbols, err := GetGoSymbols(filepath)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Symbols:", symbols)
}