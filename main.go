package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
)

var currentFile = "main.go"

type MyInterface interface {
	MyMethod()
}

func main() {
	// Find all Go files in the package directory.
	pkgPath := "."
	err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".go" && !info.IsDir() && info.Name() != currentFile {
			// Call the function in the file that returns the interface instance.
			instance := callInstanceFunc(pkgPath, path)
			if instance != nil {
				// Call a method on the interface instance.
				instance.MyMethod()
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking directory: %v", err)
	}
}

func callInstanceFunc(pkgPath, file string) MyInterface {
	// Load the file as a package.
	pkg, err := build.ImportDir(pkgPath, 0)
	if err != nil {
		fmt.Printf("Error importing package: %v", err)
		return nil
	}

	// Build the full file path.
	filePath := filepath.Join(pkg.Dir, file)

	// Parse the file to get its syntax tree.
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		fmt.Printf("Error parsing file: %v", err)
		return nil
	}

	// Find the function that returns the desired interface type.
	var instanceFunc *ast.FuncDecl
	for _, decl := range astFile.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Type.Results != nil && len(fn.Type.Results.List) == 1 {
				resultType := fn.Type.Results.List[0].Type
				if id, ok := resultType.(*ast.Ident); ok {
					if id.Name == reflect.TypeOf((*MyInterface)(nil)).Elem().Name() {
						instanceFunc = fn
						break
					}
				}
			}
		}
	}

	// Call the function and return the interface instance.
	if instanceFunc != nil {
		// Get the function name.
		funcName := instanceFunc.Name.Name
		fmt.Printf("Found function: %v", funcName)
		// Invoke instanceFunc
		t := buildFunctionType(instanceFunc)
		fmt.Println(t)
		// Locate the *types.Func object corresponding to the function.
		funcPkg, err := build.ImportDir(pkg.Dir, 0)
		if err != nil {
			panic(err)
		}
		fmt.Println(funcPkg)

		// Load the package and get a reflect.Value of the function.
		//create instance of the function
		funcValue := reflect.ValueOf(instanceFunc)

		// Call the function and return the interface instance.
		result := funcValue.Call(nil)
		return result[0].Interface().(MyInterface)
	}

	return nil
}

func buildFunctionType(funcDecl *ast.FuncDecl) reflect.Type {
	// Build the function signature.
	var paramTypes []reflect.Type
	for _, param := range funcDecl.Type.Params.List {
		paramTypes = append(paramTypes, getType(param.Type))
	}
	var resultTypes []reflect.Type
	if funcDecl.Type.Results != nil {
		for _, result := range funcDecl.Type.Results.List {
			resultTypes = append(resultTypes, getType(result.Type))
		}
	}
	funcType := reflect.FuncOf(paramTypes, resultTypes, false)

	return funcType
}

func getType(expr ast.Expr) reflect.Type {
	switch expr := expr.(type) {
	case *ast.Ident:
		return reflect.TypeOf(0)
	case *ast.StarExpr:
		return reflect.PtrTo(getType(expr.X))
	default:
		panic(fmt.Sprintf("Unsupported type: %T", expr))
	}
}
