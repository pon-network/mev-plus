package common

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	listmodules "github.com/pon-network/mev-plus/core/additionalFeatures/moduleManagement/listModules"
)

func AddModuleToModuleList(allExternalModules listmodules.ModuleList, newModule CompatibleMEVPlusCoreService) (err error) {

	source, err := os.ReadFile(moduleListFilePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		err = fmt.Errorf("error reading file: %v", err)
		return
	}

	// Create a new file set and parse the source code
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		err = fmt.Errorf("error parsing file: %v", err)
		return
	}

	// Edit the source code
	var initFunc *ast.FuncDecl
	newModuleAlias := newModule.Name

	var foundInit bool
	var foundImport bool

	for _, decl := range file.Decls {

		if !foundInit {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "init" {
				initFunc = funcDecl
				foundInit = true
			}
		}

		if !foundImport {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok == token.IMPORT {
				existingModuleCount := 1
				// check if there is an existing module with the same name
				for _, importSpec := range genDecl.Specs {
					importSpec := importSpec.(*ast.ImportSpec)
					if importSpec.Name != nil && importSpec.Name.Name == newModuleAlias {
						// There exists a module by the same alias
						splitModuleName := strings.Split(newModuleAlias, "_")
						if len(splitModuleName) == 1 {
							newModuleAlias = fmt.Sprintf("%s_%d", newModule.Variable.Name, existingModuleCount+1)
							existingModuleCount++
						} else {
							// check if the last part is a number
							num, err := strconv.Atoi(splitModuleName[len(splitModuleName)-1])
							if err == nil {
								existingModuleCount = num
								newModuleAlias = fmt.Sprintf("%s_%d", newModule.Variable.Name, existingModuleCount+1)
								existingModuleCount++
							} else {
								newModuleAlias = fmt.Sprintf("%s_%d", newModule.Variable.Name, existingModuleCount+1)
								existingModuleCount++
							}
						}
					}
				}

				// Add the new module
				genDecl.Specs = append(genDecl.Specs, &ast.ImportSpec{
					Name: &ast.Ident{
						Name: newModuleAlias,
					},
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: fmt.Sprintf("\"%s\"", strings.Split(newModule.ImportPath, "@")[0]),
					},
				})

				foundImport = true
			}
		}

		if foundImport && foundInit {
			break
		}
	}

	if !foundImport {
		err = fmt.Errorf("import statement not found")
		return
	}

	if initFunc == nil {
		fmt.Println("init function not found")
		err = fmt.Errorf("init function not found")
		return
	}

	// Check if the init function already contains an initialization for the ServiceList
	var serviceListInit *ast.AssignStmt
	for _, stmt := range initFunc.Body.List {
		if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
			if ident, ok := assignStmt.Lhs[0].(*ast.Ident); ok && ident.Name == "ServiceList" {
				serviceListInit = assignStmt
				break
			}
		}
	}

	// If the ServiceList initialization exists, append the new service initialization to it
	if serviceListInit != nil {
		// Assuming the ServiceList is initialized as a slice of services
		// Append the new service initialization to the existing initialization
		newServiceInit := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: newModuleAlias},
				Sel: &ast.Ident{Name: newModule.GenerativeFunc.Name.Name},
			},
		}

		serviceListInit.Rhs[0] = &ast.CompositeLit{
			Type: &ast.ArrayType{
				Elt: &ast.Ident{Name: "coreCommon.Service"},
			},
			Elts: append([]ast.Expr{
				newServiceInit,
			}, serviceListInit.Rhs[0].(*ast.CompositeLit).Elts...),
		}
	} else {
		// If the ServiceList initialization does not exist, create a new initialization block
		// this would only happen if there has not been a moduleList.go file before

		if len(allExternalModules.Modules) > 1 {
			// if there is more than one module, then the ServiceList initialization should be a slice of services
			// if its not there something must have gone wrong
			err = fmt.Errorf("ServiceList initialization not found, ensure the moduleList.go file is correct")
			return
		}

		newServiceInit := &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{Name: "ServiceList"},
			},
			Rhs: []ast.Expr{
				&ast.CompositeLit{
					Type: &ast.ArrayType{
						Elt: &ast.Ident{Name: "coreCommon.Service"},
					},
					Elts: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: newModuleAlias},
								Sel: &ast.Ident{Name: newModule.GenerativeFunc.Name.Name},
							},
						},
					},
				},
			},
			Tok: token.ASSIGN,
		}
		initFunc.Body.List = append([]ast.Stmt{newServiceInit}, initFunc.Body.List...)
	}

	// Create a buffer to hold the modified source code
	var buf bytes.Buffer

	// Format and print the modified source code to the buffer
	if err = format.Node(&buf, fset, file); err != nil {
		fmt.Println("Error formatting node:", err)
		err = fmt.Errorf("error formatting node: %v", err)
		return
	}

	// Write the previous version of the file to a backup file
	if err = os.WriteFile(moduleListFilePath+".bak", source, 0644); err != nil {
		fmt.Println("Error writing backup file:", err)
		err = fmt.Errorf("error writing backup file: %v", err)
		return
	}

	var errorOccured atomic.Bool
	defer func () {
		restoreErr := restoreModuleList(source, &errorOccured)
		if restoreErr != nil {
			err = fmt.Errorf("error restoring backup file: %v, after error in adding new module: %v", restoreErr, err)
		}
	}()
	// From this checkpoint a file restore would be executed on
	// any error that occurs during the process

	// Write the modified source code back to the file
	if err = os.WriteFile(moduleListFilePath, buf.Bytes(), 0644); err != nil {
		err = fmt.Errorf("error writing new moduleList.go : %v", err)
		errorOccured.Store(true)
		return
	}

	// Run go mod tidy to ensure the go.mod file is updated
	cmd := exec.Command("go", "mod", "tidy")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running go mod tidy:", err)
		err = fmt.Errorf("error running go mod tidy: %v", err)
		errorOccured.Store(true)
		return
	}

	// After it has been rewritten, check if the entire project does not have any errors
	cmd = exec.Command("go", "build", "-o", "/dev/null") // building to dev/null to discard the output
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error building project:", err)
		err = fmt.Errorf("error building project: %v", err)
		errorOccured.Store(true)
		return
	}

	return
}

func RemoveModuleFromModuleList(moduleToRemove CompatibleMEVPlusCoreService) (err error) {

	source, err := os.ReadFile(moduleListFilePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		err = fmt.Errorf("error reading file: %v", err)
		return
	}

	// Create a new file set and parse the source code
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		err = fmt.Errorf("error parsing file: %v", err)
		return
	}

	// Edit the source code
	var initFunc *ast.FuncDecl
	var moduleImport *ast.ImportSpec
	moduleToRemovePkgUrl := strings.Split(moduleToRemove.ImportPath, "@")[0]

	var foundInit bool
	var foundImport bool

	for _, decl := range file.Decls {

		if !foundInit {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "init" {
				initFunc = funcDecl
				foundInit = true
			}
		}

		if !foundImport {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok == token.IMPORT {
				for i, importSpec := range genDecl.Specs {
					importSpec := importSpec.(*ast.ImportSpec)
					if strings.EqualFold(importSpec.Path.Value, fmt.Sprintf("\"%s\"", moduleToRemovePkgUrl)) {
						genDecl.Specs = append(genDecl.Specs[:i], genDecl.Specs[i+1:]...)
						moduleImport = importSpec
						foundImport = true
					}
				}
			}
		}

		if foundImport && foundInit {
			break
		}
	}

	if !foundImport {
		err = fmt.Errorf("import statement not found")
		return
	}

	if initFunc == nil {
		fmt.Println("init function not found")
		err = fmt.Errorf("init function not found")
		return
	}

	// Check if the init function already contains an initialization for the ServiceList
	var serviceListInit *ast.AssignStmt
	for _, stmt := range initFunc.Body.List {
		if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
			if ident, ok := assignStmt.Lhs[0].(*ast.Ident); ok && ident.Name == "ServiceList" {
				serviceListInit = assignStmt
				break
			}
		}
	}

	if moduleImport.Name == nil {

		// identify package name from the root of the module path
		var retriedList bool
	getPackage:
		cmd := exec.Command("go", "list", "-f", "{{.Dir}}", "-m", moduleToRemovePkgUrl)
		output, listErr := cmd.CombinedOutput()
		if listErr != nil {
			fmt.Printf("Error executing 'go list': %v\n", err)
			if !retriedList {
				getErr := obtainPackage(moduleToRemove.ImportPath)
				if getErr != nil {
					return fmt.Errorf("failed to identify package: %v", getErr)
				}
				retriedList = true
				goto getPackage
			}
			err = fmt.Errorf("failed to identify package: %v", listErr)
			return
		}
		directoryPath := strings.TrimSpace(string(output))
		if directoryPath == "" {
			fmt.Println("Package directory not found")
			return fmt.Errorf("package directory not found")
		}

		// Detect a root Go file in the package directory
		var goFilePath string
		readErr := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
				goFilePath = path
				return filepath.SkipDir // Stop walking the directory after finding the first Go file
			}
			return nil
		})
		if readErr != nil {
			fmt.Printf("Error walking the package directory: %v\n", readErr)
			return fmt.Errorf("error walking the package directory: %v", readErr)
		}
		pkgfset := token.NewFileSet()
		pkgfile, readErr := parser.ParseFile(pkgfset, goFilePath, nil, parser.ParseComments)
		if readErr != nil {
			fmt.Printf("Error parsing the package file: %v\n", readErr)
			return fmt.Errorf("error parsing the package file: %v", readErr)
		}
		if pkgfile.Name == nil {
			fmt.Println("Package name not found")
			return fmt.Errorf("package name not found")
		}
		moduleImport.Name = &ast.Ident{Name: pkgfile.Name.Name}

	}

	var moduleRemoved bool

	// If the ServiceList initialization exists, remove the service initialization from it
	if serviceListInit != nil {
		// Assuming the ServiceList is initialized as a slice of services
		// Remove the service initialization from the existing initialization
		for i, elt := range serviceListInit.Rhs[0].(*ast.CompositeLit).Elts {
			if callExpr, ok := elt.(*ast.CallExpr); ok {
				if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if selectorExpr.X.(*ast.Ident).Name == moduleImport.Name.Name {
						serviceListInit.Rhs[0].(*ast.CompositeLit).Elts = append(serviceListInit.Rhs[0].(*ast.CompositeLit).Elts[:i], serviceListInit.Rhs[0].(*ast.CompositeLit).Elts[i+1:]...)
						moduleRemoved = true
						break
					}
				}
			}
		}
	} else {
		err = fmt.Errorf("ServiceList initialization not found, ensure the moduleList.go file is correct")
		return
	}

	if !moduleRemoved {
		err = fmt.Errorf("module not found in ServiceList to remove")
		return
	}

	// Create a buffer to hold the modified source code
	var buf bytes.Buffer

	// Format and print the modified source code to the buffer
	if err = format.Node(&buf, fset, file); err != nil {
		fmt.Println("Error formatting node:", err)
		err = fmt.Errorf("error formatting node: %v", err)
		return
	}

	// Write the previous version of the file to a backup file
	if err = os.WriteFile(moduleListFilePath+".bak", source, 0644); err != nil {
		fmt.Println("Error writing backup file:", err)
		err = fmt.Errorf("error writing backup file: %v", err)
		return
	}

	// Write the modified source code back to the file
	if err = os.WriteFile(moduleListFilePath, buf.Bytes(), 0644); err != nil {
		err = fmt.Errorf("error writing new moduleList.go : %v", err)
		return
	}

	var errorOccured atomic.Bool
	defer func () {
		restoreErr := restoreModuleList(source, &errorOccured)
		if restoreErr != nil {
			err = fmt.Errorf("error restoring backup file: %v, after error in removing module: %v", restoreErr, err)
		}
	}()
	// From this checkpoint a file restore would be executed on
	// any error that occurs

	// Run go mod tidy to ensure the go.mod file is updated
	cmd := exec.Command("go", "mod", "tidy")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running go mod tidy:", err)
		err = fmt.Errorf("error running go mod tidy: %v", err)
		errorOccured.Store(true)
		return
	}

	// After it has been rewritten, check if the entire project does not have any errors
	cmd = exec.Command("go", "build", "-o", "/dev/null") // building to dev/null to discard the output
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error building project:", err)
		err = fmt.Errorf("error building project: %v", err)
		errorOccured.Store(true)
		return
	}

	return
}

func obtainPackage(pkgURL string) (err error) {

	cmd := exec.Command("go", "get", "-d", pkgURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to obtain package: %v, output: %s", err, output)
	}

	fmt.Printf("Package obtained successfully: %s\n", pkgURL)
	return nil
}

func restoreModuleList(source []byte, executionErr *atomic.Bool) (err error) {

	// If error occurs, restore the backup file
	if executionErr.Load() {
		if restoreErr := os.WriteFile(moduleListFilePath, source, 0644); restoreErr != nil {
			fmt.Println("Error restoring backup file:", err)
			err = fmt.Errorf("error restoring backup file: %v, after error in writing new moduleList.go : %v", err, restoreErr)
			return
		}
		if removeBackUperr := os.Remove(moduleListFilePath + ".bak"); removeBackUperr != nil {
			fmt.Println("Error deleting backup file:", err)
			err = fmt.Errorf("error deleting backup file: %v, after error in writing new moduleList.go : %v", err, removeBackUperr)
			return
		}

		// Run go mod tidy to ensure the go.mod file is updated
		cmd := exec.Command("go", "mod", "tidy")
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error running go mod tidy:", err)
			err = fmt.Errorf("error running go mod tidy: %v", err)
		}
	}

	return
}
