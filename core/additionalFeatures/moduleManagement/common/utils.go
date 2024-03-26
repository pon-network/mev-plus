package common

import (
	"go/ast"
	"strings"
)

func PackageUrlToSafePath(pkgURL string) string {
	path := strings.Replace(pkgURL, "https://", "", -1)
	path = strings.Replace(path, "http://", "", -1)
	path = strings.Replace(path, "/", "_", -1)
	path = strings.Replace(path, ".", "_", -1)
	return path
}

func IdentifyAstIdentity(x ast.Expr) (*ast.Ident, string, bool) {
	switch x := x.(type) {
	case *ast.Ident:
		return x, "", true
	case *ast.SelectorExpr:
		// For a selector expression, the package name is the prefix and the type name is the suffix
		pkgName := x.X.(*ast.Ident).Name
		typeName := x.Sel.Name
		return &ast.Ident{Name: typeName}, pkgName, true
	// Handle other cases as before
	case *ast.ArrayType:
		return IdentifyAstIdentity(x.Elt)
	case *ast.MapType:
		return IdentifyAstIdentity(x.Key)
	case *ast.InterfaceType:
		return IdentifyAstIdentity(x.Methods.List[0].Type.(*ast.FuncType).Params.List[0].Type)
	case *ast.FuncType:
		return IdentifyAstIdentity(x.Params.List[0].Type)
	case *ast.StructType:
		return IdentifyAstIdentity(x.Fields.List[0].Type)
	case *ast.ChanType:
		return IdentifyAstIdentity(x.Value)
	case *ast.Ellipsis:
		return IdentifyAstIdentity(x.Elt)
	case *ast.ParenExpr:
		return IdentifyAstIdentity(x.X)
	case *ast.StarExpr:
		return IdentifyAstIdentity(x.X)
	}
	return nil, "", false
}
