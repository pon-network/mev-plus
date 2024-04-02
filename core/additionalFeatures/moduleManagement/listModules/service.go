package listmodules

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"reflect"

	"github.com/urfave/cli/v2"

	aggregator "github.com/pon-network/mev-plus/modules/block-aggregator"
	builderApi "github.com/pon-network/mev-plus/modules/builder-api"
	proxyModule "github.com/pon-network/mev-plus/modules/external-validator-proxy"
	relay "github.com/pon-network/mev-plus/modules/relay"

	moduleList "github.com/pon-network/mev-plus/moduleList"
)

type ModuleList struct {
	Modules []ModuleInfo
}

type ModuleInfo struct {
	Name           string
	Description    string
	Type           string // internal or external
	PkgReleaseDate string `json:"package_release_date,omitempty"`
	PackageUrl     string `json:"package_url,omitempty"`
}

type PackageInfo struct {
	Path    string `json:"Path"`
	Version string `json:"Version"`
	Time    string `json:"Time"`
}

// ListModules is the entry point for the list modules command
func ListModules() (ModuleList, error) {

	// Get the list of installed packages
	intalledPackages := map[string]PackageInfo{}
	cmd := exec.Command("go", "list", "-m", "-json", "all")
	output, err := cmd.StdoutPipe()
	if err != nil {
		return ModuleList{}, fmt.Errorf("Error creating stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		return ModuleList{}, fmt.Errorf("Error starting command: %v", err)
	}
	decoder := json.NewDecoder(output)
	for {
		var pkg PackageInfo
		if err := decoder.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			return ModuleList{}, fmt.Errorf("Error decoding package info: %v", err)
		}
		intalledPackages[pkg.Path] = pkg
	}
	if err := cmd.Wait(); err != nil {
		return ModuleList{}, fmt.Errorf("Error waiting for command: %v", err)
	}

	// Get the list of modules
	defaultModules := []*cli.Command{
		builderApi.NewBuilderApiService().CliCommand(),
		relay.NewRelayService().CliCommand(),
		aggregator.NewBlockAggregatorService().CliCommand(),
		proxyModule.NewExternalValidatorProxyService().CliCommand(),
	}
	externalModules := moduleList.ServiceList

	// Create the list of modules info to return
	var moduleList ModuleList
	for _, module := range defaultModules {
		moduleList.Modules = append(moduleList.Modules, ModuleInfo{
			Name:        module.Name,
			Description: module.UsageText,
			Type:        "internal",
		})
	}

	for _, module := range externalModules {
		moduleCommand := module.CliCommand()
		moduleInfo := ModuleInfo{
			Name:        moduleCommand.Name,
			Description: moduleCommand.UsageText,
			Type:        "external",
		}
		t := reflect.TypeOf(module)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		packageName := t.PkgPath()

		if pkgInfo, ok := intalledPackages[packageName]; ok {
			pkgUrl := pkgInfo.Path
			if pkgInfo.Version != "" {
				pkgUrl += "@" + pkgInfo.Version
			}
			moduleInfo.PackageUrl = pkgUrl
			moduleInfo.PkgReleaseDate = pkgInfo.Time
		}
		moduleList.Modules = append(moduleList.Modules, moduleInfo)
	}

	return moduleList, nil

}
