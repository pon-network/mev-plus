// mevPlus is the official command-line client for Mev+, the Ethereum validator proxy software.
package coreCli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bsn-eng/mev-plus/common"
	"github.com/bsn-eng/mev-plus/core"
	coreConfig "github.com/bsn-eng/mev-plus/core/config"
	moduleList "github.com/bsn-eng/mev-plus/moduleList"
	aggregator "github.com/bsn-eng/mev-plus/modules/block-aggregator/config"
	builderApi "github.com/bsn-eng/mev-plus/modules/builder-api/config"
	relay "github.com/bsn-eng/mev-plus/modules/relay/config"

	"runtime/debug"

	log "github.com/sirupsen/logrus"

	cli "github.com/urfave/cli/v2"
)

var (

	// coreConfig are flags specific to the MEV+ core and services.
	mainFlags = []cli.Flag{
		// coreConfig.PoNEnabled,
	}
)

var app = coreConfig.NewApp("the MEV+ command line interface")

func init() {
	// Initialize the CLI app and start MEV+
	app.Action = mevPlus
	app.Copyright = "Copyright 2023 BlockSwap Labs"


	var commands []*cli.Command
	commands = []*cli.Command{
		// Load default module cli commands
		builderApi.NewCommand(),
		relay.NewCommand(),
		aggregator.NewCommand(),
	}
	// Load commands for other modules
	commands = append(commands, moduleList.CommandList...)

	var moduleFlags []cli.Flag
	// then chack whether each module flag is prefixed with the module name
	if app.Metadata == nil {
		app.Metadata = make(map[string]interface{})
	}
	app.Metadata["modules"] = make([]string, 0)
	app.Metadata["moduleFlags"] = make(map[string][]cli.Flag)
	for _, cmd := range commands {

		app.Metadata["modules"] = append(app.Metadata["modules"].([]string), cmd.Name)
		for _, flag := range cmd.Flags {
			// all flags should me cmdName.FlagName

			if strings.HasPrefix(flag.Names()[0], cmd.Name) == false {
				panic(fmt.Sprintf("flag defined %s is not prefixed with module name %s", flag.Names()[0], cmd.Name))
			}

			moduleFlags = append(moduleFlags, flag)
			app.Metadata["moduleFlags"].(map[string][]cli.Flag)[cmd.Name] = append(app.Metadata["moduleFlags"].(map[string][]cli.Flag)[cmd.Name], flag)
		}
	}

	app.Flags = Merge(
		mainFlags,
		moduleFlags,
	)

}

func Run() {
	// Run the CLI app
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// mevPlus is the main entry point into the system if no special subcommand is run.
func mevPlus(ctx *cli.Context) error {

	if args := ctx.Args().Slice(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}

	core, err := makeCore(ctx)
	if err != nil {
		return err
	}
	defer core.Close()

	err = setConfigs(core, ctx)
	if err != nil {
		return err
	}

	startCore(ctx, core)

	core.Wait()

	return nil
}

func prepare(ctx *cli.Context) {

}

func makeCore(ctx *cli.Context) (*core.CoreService, error) {

	log.Info("Creating the MEV+ core and services")
	core := core.NewCoreService(ctx)

	return core, nil
}

func setConfigs(core *core.CoreService, ctx *cli.Context) error {

	log.Info("Setting the MEV+ core and services configurations")

	coreConfig := coreConfig.CoreConfig{}

	if err := setCoreConfig(ctx, &coreConfig); err != nil {
		return err
	}

	moduleFlags := make(map[string]common.ModuleFlags)

	for module, flags := range ctx.App.Metadata["moduleFlags"].(map[string][]cli.Flag) {
		for _, flag := range flags {
			for _, name := range flag.Names() {
				if ctx.IsSet(name) {
					if _, ok := moduleFlags[module]; !ok {
						moduleFlags[module] = make(common.ModuleFlags)
					}
					moduleFlags[module][name] = ctx.String(name)
				}
			}
		}

	}

	coreConfig.ModuleFlags = moduleFlags

	return core.Configure(coreConfig)
}

func setCoreConfig(ctx *cli.Context, coreConfig *coreConfig.CoreConfig) error {

	// set fields in coreConfig from ctx

	return nil
}

func startCore(ctx *cli.Context, core *core.CoreService) {

	log.Info("Starting the MEV+ core and services")

	if err := core.Start(); err != nil {
		log.WithError(err).Fatal("Failed to start the MEV+ core and services")
	}

	// Throughout the MEV+'s life cycle, listen for interrupt signals (SIGINT and SIGTERM) and
	// handles the shutdown process of the MEV+ core and services.
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)

		shutdown := func() {
			log.Info("MEV+ interrupt, shutting down...")
			go core.Close()
			for i := 10; i > 0; i-- {
				<-sigc
				if i > 1 {
					log.Warn("Already shutting down, interrupt more to panic. ", "times", i-1)
				}
			}
			debug.SetTraceback("all")
			panic("MEV+ abruptly stopped")
		}

		<-sigc
		shutdown()

	}()

	log.Info("MEV+ core and services started")

}