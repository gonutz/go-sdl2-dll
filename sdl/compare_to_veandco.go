//+build ignore

/*
This script compares the veandco to the gonutz repo and prints a list of public
declarations that differ in both packages.
Whenever the original veandco repo is updated we can use this script to find any
differences to our gonutz fork and update our package accordingly.
*/

package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	veandco := parse("veandco")
	gonutz := parse("gonutz")

	var missing, additional []string
	for _, decl := range veandco {
		if !contains(gonutz, decl) {
			missing = append(missing, decl)
		}
	}
	for _, decl := range gonutz {
		if !contains(veandco, decl) {
			additional = append(additional, decl)
		}
	}

	fmt.Println("missing items in gonutz")
	for _, d := range missing {
		fmt.Println("\t", d)
	}
	fmt.Println("additional items in gonutz")
	for _, d := range additional {
		fmt.Println("\t", d)
	}
}

func parse(owner string) []string {
	var fs token.FileSet
	path := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", owner, "go-sdl2", "sdl")
	pkgs, err := parser.ParseDir(&fs, path, func(f os.FileInfo) bool {
		name := strings.ToLower(f.Name())
		return !strings.HasSuffix(name, "_test.go")
	}, parser.AllErrors)
	if err != nil {
		panic(err)
	}
	var files []*ast.File
	for name, pkg := range pkgs {
		if name == "sdl" {
			for _, f := range pkg.Files {
				files = append(files, f)
			}
		}
	}

	config := &types.Config{
		Error:    func(e error) {},
		Importer: importer.Default(),
	}
	pkg, _ := config.Check("github.com/"+owner+"/go-sdl2/sdl", &fs, files, nil)
	scope := pkg.Scope()
	var decls []string
	for _, name := range scope.Names() {
		if ast.IsExported(name) {
			obj := scope.Lookup(name)
			decls = append(decls, name)

			typ := obj.Type()
			for _, t := range []types.Type{typ, types.NewPointer(typ)} {
				mset := types.NewMethodSet(t)
				for i := 0; i < mset.Len(); i++ {
					ith := mset.At(i)
					if ast.IsExported(ith.Obj().Name()) {
						s := fmt.Sprint(ith)
						s = strings.Replace(s, "github.com/"+owner+"/go-sdl2/sdl.", "", -1)
						decls = append(decls, s)
					}
				}
			}
		}
	}
	return decls
}

func contains(list []string, d string) bool {
	for i := range list {
		if list[i] == d {
			return true
		}
	}
	return false
}
