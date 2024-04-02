package common

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	coreCommon "github.com/pon-network/mev-plus/core/common"
)

// InterfaceChecker checks if a package has a struct that meets a specified interface.
type InterfaceChecker struct {
	InterfaceName string
	PkgPath       string
	Methods       []MethodInfo
}

// MethodInfo stores information about a method, including its name, argument types, and return types.
type MethodInfo struct {
	Name    string
	Args    []ArgInfo
	Returns []OutputInfo
}

type ArgInfo struct {
	Type    reflect.Type
	Kind    reflect.Kind
	PkgPath string
}

type OutputInfo struct {
	Type    reflect.Type
	Kind    reflect.Kind
	PkgPath string
}

// Dynamically initialize the InterfaceChecker based on the MEV Plus Service interface
func init() {
	serviceType := reflect.TypeOf((*coreCommon.Service)(nil)).Elem()
	methods := make([]MethodInfo, serviceType.NumMethod())
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)

		numOut := method.Type.NumOut()
		outputs := make([]OutputInfo, numOut)
		for j := 0; j < numOut; j++ {
			output := method.Type.Out(j)
			pkgPath := output.PkgPath()

			// if it is a pointer, get the underlying type
			if output.Kind() == reflect.Ptr {
				outputElem := output.Elem()
				pkgPath = outputElem.PkgPath()
			}
			outputs[j] = OutputInfo{
				Type:    output,
				PkgPath: pkgPath,
				Kind:    output.Kind(),
			}
		}

		numIn := method.Type.NumIn()
		args := make([]ArgInfo, numIn)
		for j := 0; j < numIn; j++ {
			arg := method.Type.In(j)
			pkgPath := arg.PkgPath()

			// if it is a pointer, get the underlying type
			if arg.Kind() == reflect.Ptr {
				argElem := arg.Elem()
				pkgPath = argElem.PkgPath()
			}
			args[j] = ArgInfo{
				Type:    arg,
				Kind:    arg.Kind(),
				PkgPath: pkgPath,
			}

		}

		methodInfo := MethodInfo{
			Name:    method.Name,
			Args:    args,
			Returns: outputs,
		}

		methods[i] = methodInfo
	}

	IC = &InterfaceChecker{
		InterfaceName: serviceType.Name(),
		PkgPath:       serviceType.PkgPath(),
		Methods:       methods,
	}

}

// This is done this way so as if the MEV Plus Core Service is ever
// updated, the InterfaceChecker will automatically reflect the changes of the spec

var IC *InterfaceChecker

func (ic *InterfaceChecker) ImplementsInterface(file *ast.File) (variable *ast.Ident, identifiedStruct *ast.Ident, ok bool) {
	// To check if the file which has the identified struct implements the interface (MEV Plus Core Service)

	numberOfSupportedMethods := len(ic.Methods)

	var identifiedReceivers map[*ast.Ident]int = make(map[*ast.Ident]int)

	// **IMPORTANT**
	// The file may contain other functions and methods that have nothing to do with the interface
	// However if the file does contain a struct that implements the interface, it would have all
	// should have all the methods of the interface declared in the same file as the struct

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Check if the function has a receiver
		// meaning the function is a method
		if funcDecl.Recv == nil {
			continue
		}

		// The receiver must be a single struct which would be further checked
		// to see if the receiver is of the MEV Plus Core Service interface
		if len(funcDecl.Recv.List) != 1 {
			continue
		}

		receiverType := funcDecl.Recv.List[0].Type
		var receiverTypeIdent *ast.Ident // This holds the potential receiver type identifier
		switch t := receiverType.(type) {
		case *ast.Ident:
			// Direct identifier, i.e., "func (s MEVPlusService) MethodName()"
			receiverTypeIdent = t
		case *ast.StarExpr:
			// Pointer to a type, i.e., "func (s *MEVPlusService) MethodName()"
			if ident, ok := t.X.(*ast.Ident); ok {
				receiverTypeIdent = ident
			}
		default:
			// fmt.Println("Unsupported receiver type: ", reflect.TypeOf(t))
			continue
		}

		// Found a method with a matching receiver, now check if it matches one of the methods in the InterfaceChecker
		importPaths := parseImports(file)
		for _, methodInfo := range ic.Methods {
			if funcDecl.Name.Name == methodInfo.Name {

				if !ic.compareMethodSignature(funcDecl, methodInfo, importPaths) {
					// the method signature does not match the expected signature
					continue
				}

				// Method signature matches the expected signature

				if len(identifiedReceivers) == 0 {
					// If there are no identified receivers, add the receiver to the identifiedReceivers
					identifiedReceivers[receiverTypeIdent] = 1
				} else {
					for identifiedReceiver, numberOfSupportedMethodsFound := range identifiedReceivers {
						if identifiedReceiver.Name == receiverTypeIdent.Name {
							identifiedReceivers[identifiedReceiver] = numberOfSupportedMethodsFound + 1
						}
					}
				}

			}
		}
	}

	// If we've checked all methods and found them all, the struct implements the interface
	for identifiedReceiver, numberOfSupportedMethodsFound := range identifiedReceivers {
		if numberOfSupportedMethodsFound == numberOfSupportedMethods {

			s := fmt.Sprintf(
				"\nThe package contains receiver `%s` of type `%s` which has %d/%d methods that implement the MEV Plus Core Service interface\n",
				file.Name.Name,
				identifiedReceiver.Name,
				numberOfSupportedMethodsFound,
				numberOfSupportedMethods,
			)
			fmt.Println(s)

			return file.Name, identifiedReceiver, true
		}
	}

	return nil, nil, false
}

func parseImports(file *ast.File) map[string]string {
	importPaths := make(map[string]string)
	for _, imp := range file.Imports {
		if imp.Name != nil {
			// If the import has an alias, use the alias as the key
			name := imp.Name.Name
			if name == "_" {
				// If the alias is an underscore, ignore it
				continue
			}
			name = strings.ReplaceAll(name, "\"", "")
			name = strings.ReplaceAll(name, "'", "")
			path := imp.Path.Value
			path = strings.ReplaceAll(path, "\"", "")
			path = strings.ReplaceAll(path, "'", "")
			importPaths[name] = path
		} else {
			// Otherwise, use the last part of the import path as the key
			path := imp.Path.Value
			path = strings.ReplaceAll(path, "\"", "")
			path = strings.ReplaceAll(path, "'", "")
			lastSlash := strings.LastIndex(path, "/")
			if lastSlash != -1 {
				importPaths[path[lastSlash+1:]] = path
			}
		}
	}
	return importPaths
}

func (ic *InterfaceChecker) compareMethodSignature(funcDecl *ast.FuncDecl, methodInfo MethodInfo, importPaths map[string]string) bool {
	// Check if the function conforms to the expected method specification

	// Arguments Check
	if len(funcDecl.Type.Params.List) == 0 && len(methodInfo.Args) == 0 {
		// If there is no arguments and no expected arguments, then the method signature matches on the arguments
		// fmt.Println(fmt.Sprintf(
		// 	"Method `%s` from `%s` has no arguments and no expected arguments",
		// 	funcDecl.Name.Name,
		// 	funcDecl.Recv.List[0].Type,
		// ))
	} else {
		// Check the number of arguments
		if len(funcDecl.Type.Params.List) != len(methodInfo.Args) {
			return false
		}
		for i, arg := range funcDecl.Type.Params.List {
			var argType *ast.Ident
			argType, argImportPrefix, ok := IdentifyAstIdentity(arg.Type)
			if !ok {
				// fmt.Println("Unsupported argument type: ", reflect.TypeOf(arg.Type))
				return false
			}

			expectedArgElem := methodInfo.Args[i].Type
			if methodInfo.Args[i].Kind == reflect.Ptr {
				// If the expected argument is a pointer, get the underlying type
				expectedArgElem = methodInfo.Args[i].Type.Elem()
			}

			// Compare the argument type name
			if argType.Name != expectedArgElem.Name() {
				// No need to check further argument types if one does not match
				return false
			}

			// Compare the argument type package path
			if methodInfo.Args[i].PkgPath != "" && argImportPrefix == "" {
				// If the expected argument has a package path, means the argument must be imported
				// if the argImportPrefix is empty, it means the argument failed to meet this spec and is
				// a local type
				return false
			} else if argImportPrefix != "" && methodInfo.Args[i].PkgPath != "" {
				// If the argument is imported, check if the package path matches
				argPkgPath := importPaths[argImportPrefix]
				if argPkgPath != methodInfo.Args[i].PkgPath {
					// If the package path does not match, the argument does not of the right imported type
					return false
				}
			}

			// argPkgFound := ""
			// if methodInfo.Args[i].PkgPath != "" {
			// 	argPkgFound = fmt.Sprintf(" from `%s`", methodInfo.Args[i].PkgPath)
			// }
			// fmt.Println(fmt.Sprintf(
			// 	"Method `%s` from `%s` argument type `%s` matches expected `%s`%s at expected args index `%d`",
			// 	funcDecl.Name.Name,
			// 	funcDecl.Recv.List[0].Type,
			// 	argType.Name,
			// 	expectedArgElem.Name(),
			// 	argPkgFound,
			// 	i,
			// ))
		}
	}

	// Return Check
	if len(funcDecl.Type.Results.List) == 0 && len(methodInfo.Returns) == 0 {
		// If there is no return values and no expected return values, then the method signature matches on the return values
		// fmt.Println(fmt.Sprintf(
		// 	"Method `%s` from `%s` has no return values and no expected return values",
		// 	funcDecl.Name.Name,
		// 	funcDecl.Recv.List[0].Type,
		// ))
	} else {
		// Check the number of return values
		if len(funcDecl.Type.Results.List) != len(methodInfo.Returns) {
			return false
		}
		for i, result := range funcDecl.Type.Results.List {
			var returnType *ast.Ident
			returnType, returnImportPrefix, ok := IdentifyAstIdentity(result.Type)
			if !ok {
				// fmt.Println("Unsupported return type: ", reflect.TypeOf(result.Type))
				return false
			}

			expectedReturnElem := methodInfo.Returns[i].Type
			if methodInfo.Returns[i].Kind == reflect.Ptr {
				// If the expected return value is a pointer, get the underlying type
				expectedReturnElem = methodInfo.Returns[i].Type.Elem()
			}

			// Compare the return value type name
			if returnType.Name != expectedReturnElem.Name() {
				// No need to check further return types if one does not match
				return false
			}

			// Compare the return value type package path
			if methodInfo.Returns[i].PkgPath != "" && returnImportPrefix == "" {
				// If the expected return value has a package path, means the return value must be imported
				// if the returnImportPrefix is empty, it means the return value failed to meet this spec and is
				// a local type
				return false
			} else if returnImportPrefix != "" && methodInfo.Returns[i].PkgPath != "" {
				// If the return value is imported, check if the package path matches
				returnPkgPath := importPaths[returnImportPrefix]
				if returnPkgPath != methodInfo.Returns[i].PkgPath {
					// If the package path does not match, the return value does not of the right imported type
					return false
				}
			}

			// returnPkgFound := ""
			// if methodInfo.Returns[i].PkgPath != "" {
			// 	returnPkgFound = fmt.Sprintf(" from `%s`", methodInfo.Returns[i].PkgPath)
			// }
			// fmt.Println(fmt.Sprintf(
			// 	"Method `%s` from `%s` return type `%s` matches expected `%s`%s at expected returns index `%d`",
			// 	funcDecl.Name.Name,
			// 	funcDecl.Recv.List[0].Type,
			// 	returnType.Name,
			// 	expectedReturnElem.Name(),
			// 	returnPkgFound,
			// 	i,
			// ))
		}
	}

	return true

}