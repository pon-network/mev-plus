package common

import (
	"go/ast"
)

type CompatibleMEVPlusCoreService struct {
	Name           string        // the name of the module
	Variable       *ast.Ident    // the variable that holds the struct
	StructDef      *ast.Ident    // the struct definition
	ImportPath     string        // the import path of the struct
	FilePath       string        // the go file path of the struct
	PkgDir         string        // the package directory
	GenerativeFunc *ast.FuncDecl // the function that generates the struct
}