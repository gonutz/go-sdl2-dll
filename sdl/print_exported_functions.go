//+build ignore

/*
This script prints all exported functions declared in the sdl package.
*/

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"sort"
)

func main() {
	funcs := make(map[string]bool)
	checkFile("sdl_windows.go", funcs)
	checkFile("sdl_windows_386.go", funcs)
	checkFile("sdl_windows_amd64.go", funcs)
	var list []string
	for name := range funcs {
		list = append(list, name)
	}
	sort.Strings(list)
	for _, name := range list {
		fmt.Println(name)
	}
}

func checkFile(path string, funcs map[string]bool) {
	// load the SDL2 code here and parse it
	code, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var fs token.FileSet
	astFile, err := parser.ParseFile(&fs, "", code, 0)
	if err != nil {
		panic(err)
	}

	for _, decl := range astFile.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok {
			if ast.IsExported(f.Name.Name) {
				name := f.Name.Name
				if f.Recv != nil && f.Recv.List[0] != nil {
					name += " (" + typeName(f.Recv.List[0].Type) + ")"
				}
				funcs[name] = true
			}
		}
	}
}

func typeName(ex ast.Expr) string {
	switch e := ex.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return typeName(e.X)
	default:
		panic(fmt.Sprintf("unknown typeName of %#v\n", ex))
	}
}
