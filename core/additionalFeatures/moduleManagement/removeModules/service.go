package removemodules

import (
	"fmt"
	"strings"

	"github.com/pon-network/mev-plus/core/additionalFeatures/moduleManagement/common"
	listmodules "github.com/pon-network/mev-plus/core/additionalFeatures/moduleManagement/listModules"
)

func RemoveModule(module string) error {

	// Check if the module is already installed
	modules, err := listmodules.ListModules()
	if err != nil {
		return fmt.Errorf("failed to list modules: %v", err)
	}

	var moduleToRemove listmodules.ModuleInfo

	for _, existingModule := range modules.Modules {

		if (existingModule.Name == module) || strings.EqualFold(strings.Split(existingModule.PackageUrl, "@")[0], module) {
			// case sensitive check for module name
			// case insensitive check for module url

			if existingModule.Type != "external" {
				return fmt.Errorf("module is not an external module: %v", module)
			}

			moduleToRemove = existingModule
			break
		}

	}

	if moduleToRemove.Name == "" || moduleToRemove.PackageUrl == "" {
		return fmt.Errorf("module is not installed: %v", module)
	}

	// removing the module from the list of modules. Operation requires just the module name and package url
	err = common.RemoveModuleFromModuleList(common.CompatibleMEVPlusCoreService{
		Name:       moduleToRemove.Name,
		ImportPath: moduleToRemove.PackageUrl,
	})
	if err != nil {
		return fmt.Errorf("failed to remove module from module list: %v", err)
	}

	fmt.Printf("\nModule removed: %v [%v]\n", moduleToRemove.Name, moduleToRemove.PackageUrl)

	return nil
}
