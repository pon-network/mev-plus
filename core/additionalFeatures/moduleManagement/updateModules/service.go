package updatemodules

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/pon-network/mev-plus/core/additionalFeatures/moduleManagement/common"
	listmodules "github.com/pon-network/mev-plus/core/additionalFeatures/moduleManagement/listModules"
)

func UpdateModule(moduleUrl string) (err error) {

	moduleUrl = strings.ToLower(strings.TrimSpace(moduleUrl))

	// Check if the module is already installed
	modules, err := listmodules.ListModules()
	if err != nil {
		err = fmt.Errorf("failed to list modules: %v", err)
		return
	}

	var foundModule listmodules.ModuleInfo
	var existingExternalModules listmodules.ModuleList

	for _, module := range modules.Modules {
		if module.Type != "external" {
			continue
		}

		if strings.EqualFold(strings.Split(module.PackageUrl, "@")[0], strings.Split(moduleUrl, "@")[0]) {
			foundModule = module
			if len(strings.Split(moduleUrl, "@")) == 1 {
				moduleUrl = fmt.Sprintf("%s@latest", moduleUrl)
			}
		} else if strings.EqualFold(moduleUrl, module.Name) {
			foundModule = module
			// module name was provided to update then take the existing package url
			// and update to the latest version
			moduleUrl = fmt.Sprintf("%s@latest", strings.Split(module.PackageUrl, "@")[0])
		}
		existingExternalModules.Modules = append(existingExternalModules.Modules, module)
	}

	if foundModule.Name == "" || foundModule.PackageUrl == "" {
		err = fmt.Errorf("module is not installed: %v", strings.Split(moduleUrl, "@")[0])
		return
	} else if strings.EqualFold(foundModule.PackageUrl, moduleUrl) {
		fmt.Printf("Module already updated to the provided version: %s\n", moduleUrl)
		return
	}

	var errorOccured atomic.Bool
	defer func() {
		restoreErr := restorePackageVersion(foundModule.PackageUrl, &errorOccured)
		if restoreErr != nil {
			err = fmt.Errorf("error restoring package version: %v, after error in updating module: %v", restoreErr, err)
		}
	}()

	// Obtain the package and check if it meets the core service requirements
	service, err := obtainAndCheckPackage(moduleUrl, existingExternalModules)
	if err != nil {
		err = fmt.Errorf("failed to obtain and check package: %v", err)
		errorOccured.Store(true)
		return
	}

	fmt.Printf("Compatible service found: %v (%v) [%v]\n", service.Variable.Name, service.StructDef.Name, service.ImportPath)

	// Run go mod tidy to ensure the go.mod and go.sum file is updated
	// since the obtain and check step already downloads the package
	// and adds it to the go.mod file
	cmd := exec.Command("go", "mod", "tidy")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running go mod tidy:", err)
		err = fmt.Errorf("error running go mod tidy: %v", err)
		errorOccured.Store(true)
		return
	}

	cmd = exec.Command("go", "build", "-o", "/dev/null") // building to dev/null to discard the output
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error building project:", err)
		err = fmt.Errorf("error building project: %v", err)
		errorOccured.Store(true)
		return
	}

	fmt.Printf("\nModule %v updated to %v\n", foundModule.Name, strings.Split(moduleUrl, "@")[1])

	return nil
}

func restorePackageVersion(pkgURL string, errorOccured *atomic.Bool) (err error) {
	if errorOccured.Load() {
		cmd := exec.Command("go", "get", "-d", pkgURL)
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error restoring original package version:", err)
		}
	}
	return
}

func obtainAndCheckPackage(pkgURL string, existingExternalPackages listmodules.ModuleList) (service common.CompatibleMEVPlusCoreService, err error) {
	// This would obtain the package if it exists and
	// check if it meets the core service requirements

	err = obtainPackage(pkgURL)
	if err != nil {
		err = fmt.Errorf("failed to obtain package: %v", err)
		return
	}

	service, err = checkPackage(pkgURL, existingExternalPackages)
	if err != nil {
		err = fmt.Errorf("package check failed: %v", err)
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

func checkPackage(pkgURL string, existingPackageUrls listmodules.ModuleList) (compatibleService common.CompatibleMEVPlusCoreService, err error) {

	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", "-m", pkgURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("failed to find package location: %v, output: %s", err, output)
		return
	}
	pkgDir := strings.TrimSpace(string(output))

	fset := token.NewFileSet()
	err = filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// **IMPORTANT**
		// This checker assumes the package is in the same directory as the module
		// If there is a service that implements the MEV Plus core service interface
		// all service methods MUST be in the same file as the struct definition

		rerunInspectionForFile := false
	runInspection:

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				fmt.Println("Error parsing file:", err)
				return err
			}

			// Visit all declarations in the file
			ast.Inspect(file, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					if compatibleService.StructDef != nil && compatibleService.GenerativeFunc == nil {
						// Check if the function generates the struct
						if x.Type.Results != nil && len(x.Type.Results.List) == 1 {
							if ident, _, ok := common.IdentifyAstIdentity(x.Type.Results.List[0].Type); ok {
								if ident.Name == compatibleService.StructDef.Name { // a case sensitive check since different structs can have the same name but in different cases
									compatibleService.GenerativeFunc = x
									fmt.Printf("Function %s from %s generates the struct %s.\n", x.Name.Name, file.Name.Name, compatibleService.StructDef.Name)
									return false // Stop the inspection as the function is found
								}
							}
						}
					}
				case *ast.TypeSpec:
					// Check if the type is a struct
					if _, ok := x.Type.(*ast.StructType); ok {
						// Check if the struct implements the interface
						if compatibleService.StructDef == nil {
							// Compatible service not found yet, check if the struct implements the interface
							variable, identifiedStruct, ok := common.IC.ImplementsInterface(file)
							if ok {
								compatibleService = common.CompatibleMEVPlusCoreService{
									Name:       variable.Name,
									Variable:   variable,
									StructDef:  identifiedStruct,
									ImportPath: pkgURL,
									FilePath:   pkgURL + strings.Replace(path, pkgDir, "", 1),
								}

								// Use the variable name as the module name
								// for the installation. This does not affect the module official
								// name, this is just used for the moduleList definitions
								existingCount := 1
								for _, module := range existingPackageUrls.Modules {
									existingName := module.Name
									if existingName == compatibleService.Name {
										compatibleService.Name = fmt.Sprintf("%s_%d", compatibleService.Variable.Name, existingCount+1)
										existingCount++
									}
								}

								return false // Stop the inspection as the interface is implemented
							}
						}
					}
				}
				return true
			})

			if compatibleService.StructDef != nil && compatibleService.GenerativeFunc == nil && !rerunInspectionForFile {
				// Found the struct that implements the interface, now find the function that generates the struct
				rerunInspectionForFile = true
				goto runInspection
			}

		}
		return nil
	})

	if err != nil {
		fmt.Println("Error analyzing the package directory:", err)
		err = fmt.Errorf("failed to analyze package directory: %v", err)
		return
	}

	if compatibleService.StructDef == nil {
		fmt.Printf("Compatible service not found for %s\n", pkgURL)
		err = fmt.Errorf("compatible service not found for %s", pkgURL)
	}

	return
}
