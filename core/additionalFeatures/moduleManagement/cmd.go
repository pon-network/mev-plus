package moduleManagement

import (
	"fmt"
	"strings"

	"github.com/pon-network/mev-plus/cmd/utils"
	installmodules "github.com/pon-network/mev-plus/core/additionalFeatures/moduleManagement/installModules"
	listmodules "github.com/pon-network/mev-plus/core/additionalFeatures/moduleManagement/listModules"
	removemodules "github.com/pon-network/mev-plus/core/additionalFeatures/moduleManagement/removeModules"
	updatemodules "github.com/pon-network/mev-plus/core/additionalFeatures/moduleManagement/updateModules"
	cli "github.com/urfave/cli/v2"
)

func NewCommand() *cli.Command {

	return &cli.Command{
		Name:      "modules",
		Action:    modules,
		Usage:     "Manage the MEV Plus modules",
		UsageText: "The modules command is used to manage the MEV Plus modules",
		Category:  utils.CoreCategory,
		Flags:     moduleManagementFlags(),
	}
}

func moduleManagementFlags() []cli.Flag {
	return []cli.Flag{
		// The below flags can be used only one at a time
		// *required*
		InstallModuleFlags,
		RemoveModuleFlags,
		UpdateModuleFlags,
		GetModulesFlags,
	}
}

// The below function is the entry point for the modules command
func modules(ctx *cli.Context) error {

	if args := ctx.Args().Slice(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}

	requiredFlagsTrack := map[string]bool{
		strings.Join(InstallModuleFlags.Names(), ","): false,
		strings.Join(RemoveModuleFlags.Names(), ","):  false,
		strings.Join(UpdateModuleFlags.Names(), ","):  false,
		strings.Join(GetModulesFlags.Names(), ","):    false,
	}
	requiredFlagSet := ""

	for _, flag := range ctx.Command.Flags {
		for _, name := range flag.Names() {
			if ctx.IsSet(name) {
				if _, ok := requiredFlagsTrack[strings.Join(flag.Names(), ",")]; ok {
					if requiredFlagSet != "" && requiredFlagSet != strings.Join(flag.Names(), ",") {
						return fmt.Errorf("multiple flags set, only one flag can be set at a time")
					}
					requiredFlagsTrack[strings.Join(flag.Names(), ",")] = true
					requiredFlagSet = strings.Join(flag.Names(), ",")
				}
			}
		}
	}

	// If no flags are set, then print the usage
	if requiredFlagSet == "" {
		fmt.Println("Either call the command with one of the below flags:")
		fmt.Println("  --install <package-url>")
		fmt.Println("  --remove <package-name>")
		fmt.Println("  --update <package-name>")
		fmt.Println("  --list")
		fmt.Println("Or call the command with the --help flag to see the usage")
	} else {

		// If the required flags are set, then call the respective function
		if ctx.Bool("list") {
			err := listModules(ctx)
			if err != nil {
				return err
			}
		} else if ctx.String("install") != "" {
			err := installModule(ctx)
			if err != nil {
				return err
			}
		} else if ctx.String("remove") != "" {
			err := removeModule(ctx)
			if err != nil {
				return err
			}
		} else if ctx.String("update") != "" {
			err := updateModule(ctx)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func listModules(_ *cli.Context) error {
	modules, err := listmodules.ListModules()
	if err != nil {
		return err
	}

	fmt.Println("Available Modules:")
	fmt.Println("====================================")
	fmt.Println()
	for _, module := range modules.Modules {
		fmt.Println("Name:", module.Name)
		fmt.Println("Description:", module.Description)
		fmt.Println("Type:", module.Type)
		if module.PackageUrl != "" {
			fmt.Println("Package URL:", module.PackageUrl)
		}
		if module.PkgReleaseDate != "" {
			fmt.Println("Package Release Date:", module.PkgReleaseDate)
		}
		fmt.Println()
	}
	return nil
}

func installModule(ctx *cli.Context) error {

	packageURL := ctx.String("install")
	err := installmodules.InstallModule(packageURL)
	if err != nil {
		return err
	}
	return nil
}

func removeModule(ctx *cli.Context) error {

	module := ctx.String("remove")
	err := removemodules.RemoveModule(module)
	if err != nil {
		return err
	}
	return nil
}

func updateModule(ctx *cli.Context) error {

	module := ctx.String("update")
	err := updatemodules.UpdateModule(module)
	if err != nil {
		return err
	}
	return nil
}
