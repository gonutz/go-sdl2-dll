//+build ignore

/*
This script helps to find typos in the code. Wrapping SDL2 means converting a
lot of functions to calls to the DLL. During the process we might accidentally
call the wrong functions.

This script is closely related to the structure of our library, it parses it and
makes sure we do not have any of these typical typos in our code.
The output of this script should be empty, only if it finds an error, it will
print it to stdout.
*/

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

func main() {
	checkFile("sdl_windows.go")
	checkFile("sdl_windows_386.go")
	checkFile("sdl_windows_amd64.go")
}

func checkFile(path string) {
	// load the SDL2 code here and parse it
	funcs, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var fs token.FileSet
	astFile, err := parser.ParseFile(&fs, "", funcs, 0)
	if err != nil {
		panic(err)
	}

	// The SDL2.dll functions are loaded into variables like this:
	//		clearQueuedAudio = dll.NewProc("SDL_ClearQueuedAudio")
	// Make sure that the variable names all correspond to the loaded functions.
	type dllFuncVar struct {
		varName    string
		loadedFunc string
	}
	var dllFuncVars []dllFuncVar
	// we gather all declarations where a variable is assigned like
	// 	<varName> = dll.NewProc(<loadedFunc as string literal>)
	for _, decl := range astFile.Decls {
		if spec, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range spec.Specs {
				if spec, ok := spec.(*ast.ValueSpec); ok {
					if len(spec.Names) == 1 {
						dllFuncName := spec.Names[0].Name
						if len(spec.Values) == 1 {
							if call, ok := spec.Values[0].(*ast.CallExpr); ok {
								if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
									if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "dll" {
										if len(call.Args) == 1 {
											if s, ok := call.Args[0].(*ast.BasicLit); ok && s.Kind == token.STRING {
												dllFuncVars = append(dllFuncVars, dllFuncVar{
													varName:    dllFuncName,
													loadedFunc: s.Value,
												})
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	// now check everything we found
	for _, fv := range dllFuncVars {
		v := strings.ToLower(fv.varName)
		f := strings.TrimPrefix(strings.ToLower(trimQuotes(fv.loadedFunc)), "sdl_")
		// there are functions SDL_Error and SDL_Init which cannot be turned
		// into variables error and init because these are already used, thus
		// they are called sdlError and sdlInit. We need to account for them.
		if !(v == f || v == "sdl"+f) {
			// we have a mismatch!
			fmt.Println(path, fv.varName, "loads", fv.loadedFunc)
		}
	}

	// In the code we have functions that look like this:
	// 	func ClearQueuedAudio(dev AudioDeviceID) {
	// 		clearQueuedAudio.Call(uintptr(dev))
	// 	}
	// We want to make sure that the API function calls the variable with the
	// same name.
	//
	// The thing here is that we also have methods like this:
	// 	func (renderer *Renderer) CopyEx(...) error {
	// 		ret, _, _ := renderCopyEx.Call(...)
	// 	}
	// so we can have the type - or in this case a modified version of the type
	// - as the DLL function that is called.
	//
	// Another one is:
	// 	func (window *Window) Destroy() error {
	// 		destroyWindow.Call(uintptr(unsafe.Pointer(window)))
	// 	}
	// where the order of the function name and type is reversed.
	//
	// Let's go!
	type dllCall struct {
		apiFunc  string   // e.g. LoadWAV for  func LoadWAV(file string) []byte
		typeName string   // empty or e.g. Window for  func (w *Window) SetTitle()
		dllCalls []string // all calls, there can be any number of them
	}
	var dllCalls []dllCall
	for _, decl := range astFile.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok {
			if ast.IsExported(f.Name.Name) {
				var call dllCall
				call.apiFunc = f.Name.Name
				// we have an exported function, this is part of the public API
				// so we check it
				if f.Recv != nil && f.Recv.List[0] != nil {
					call.typeName = typeName(f.Recv.List[0].Type)
				}
				w := dllCallFinder{}
				ast.Walk(&w, f.Body)
				call.dllCalls = w.calls
				dllCalls = append(dllCalls, call)
			}
		}
	}
	for _, c := range dllCalls {
		if len(c.dllCalls) == 0 {
			// functions that have no calls to DLLs are helper functions like
			// Btoi or Event.GetType()
			continue
		}
		var ok bool
		for _, dllCall := range c.dllCalls {
			// convert both to lower case for comparison, we only care if the
			// words are right
			apiFunc := strings.ToLower(c.apiFunc)
			call := strings.ToLower(dllCall)
			typeName := strings.ToLower(c.typeName)
			// account for GL functions like GLDeleteContext which calls
			// gl_DeleteContext
			call = strings.Replace(call, "_", "", -1)

			if apiFunc == call {
				ok = true
			}

			// account for sdlError and sdlInit
			if "sdl"+apiFunc == call {
				ok = true
			}

			// account for getters like GetEventState which calls eventState
			if apiFunc == "get"+call {
				ok = true
			}

			// account for methods such as AudioStream.Available  which calls
			// audioStreamAvailable by removing the type name from the function
			call = strings.Replace(call, typeName, "", 1)
			if apiFunc == call {
				ok = true
			}

			// account for method getters like GameController.Attached which
			// calls gameControllerGetAttached
			call = strings.Replace(call, "get", "", 1)
			if apiFunc == call {
				ok = true
			}

			// methods of the Renderer start with "render" instead of
			// "renderer", e.g. Renderer.Clear calls renderClear so account
			// for that
			if typeName == "renderer" {
				call = strings.Replace(call, "render", "", 1)
				if apiFunc == call {
					ok = true
				}
				if apiFunc == "get"+call {
					ok = true
				}
			}

			// here are some special cases which are not covered by the above rules
			special := [][3]string{
				{"PixelFormat", "Free", "freeFormat"},
				{"RWops", "Close", "rwClose"},
				{"RWops", "Free", "freeRW"},
				{"Sem", "Destroy", "destroySemaphore"},
				{"SharedObject", "Unload", "unloadObject"},
				{"Texture", "UpdateRGBA", "updateTexture"},
			}
			for _, s := range special {
				if c.typeName == s[0] && c.apiFunc == s[1] && dllCall == s[2] {
					ok = true
				}
			}
		}
		if !ok {
			fmt.Println(path, c.typeName, c.apiFunc, c.dllCalls)
		}
	}
}

func trimQuotes(s string) string {
	return strings.Replace(s, `"`, "", -1)
}

type dllCallFinder struct {
	calls []string
}

func (w *dllCallFinder) Visit(node ast.Node) ast.Visitor {
	if call, ok := node.(*ast.CallExpr); ok {
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "Call" {
				if id, ok := sel.X.(*ast.Ident); ok {
					w.calls = append(w.calls, id.Name)
				}
			}
		}
	}
	return w
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
