package moduleManagement

import (
	"github.com/pon-network/mev-plus/cmd/utils"
	"github.com/urfave/cli/v2"
)

var (
	GetModulesFlags = &cli.BoolFlag{
		Name:               "list",
		Aliases:            []string{"l"},
		Usage:              "List all available modules",
		Category:           utils.CoreCategory,
		DisableDefaultText: true,
	}

	InstallModuleFlags = &cli.StringFlag{
		Name:     "install",
		Aliases:  []string{"i"},
		Usage:    "Install a module by passing its package url",
		Category: utils.CoreCategory,
	}

	RemoveModuleFlags = &cli.StringFlag{
		Name:     "remove",
		Aliases:  []string{"r"},
		Usage:    "Remove a module by passing its installed package name or package url",
		Category: utils.CoreCategory,
	}

	UpdateModuleFlags = &cli.StringFlag{
		Name:     "update",
		Aliases:  []string{"u"},
		Usage:    "Update a module to its latest version by passing its installed package name or package url, you can also update to a specific version by passing the package url with the version",
		Category: utils.CoreCategory,
	}
)
