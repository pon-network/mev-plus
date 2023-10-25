package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pon-pbs/mev-plus/common"
	coreCommon "github.com/pon-pbs/mev-plus/core/common"
	"github.com/pon-pbs/mev-plus/core/config"
	moduleList "github.com/pon-pbs/mev-plus/moduleList"

	blockaggregator "github.com/pon-pbs/mev-plus/modules/block-aggregator"
	builderapi "github.com/pon-pbs/mev-plus/modules/builder-api"
	"github.com/pon-pbs/mev-plus/modules/relay"

	cli "github.com/urfave/cli/v2"

	log "github.com/sirupsen/logrus"
)

const (
	initializingState = iota
	runningState
	closedState
)

// CoreService represents the ccore service that handles events and notifies attached services.
type CoreService struct {
	moduleRegistry  coreCommon.ModuleRegistry
	moduleChannels  map[string]coreCommon.ModuleCommChannels
	moduleClientIds map[string]string
	config          config.CoreConfig

	idgen func() string

	coreClient         *coreCommon.Client
	coreClientChannels coreCommon.ModuleCommChannels

	stop chan struct{} // Channel to wait for termination notifications

	state         int // Tracks state of the core service
	lock          sync.Mutex
	startStopLock sync.Mutex // Start/Stop are protected by an additional lock
}

// NewCoreService creates a new instance of the CoreService.
func NewCoreService(ctx *cli.Context) *CoreService {

	core := &CoreService{
		stop:            make(chan struct{}),
		idgen:           common.NewID,
		moduleChannels:  make(map[string]coreCommon.ModuleCommChannels),
		moduleClientIds: make(map[string]string),
	}

	// Register the default modules
	defaultServices, err := core.defaultServices()
	if err != nil {
		panic(err)
	}
	for _, service := range defaultServices {
		if err := core.moduleRegistry.RegisterName(service.Name(), service); err != nil {
			panic(err)
		}
	}

	// Register the additional modules with the core service
	for _, service := range moduleList.ServiceList {
		if err := core.moduleRegistry.RegisterName(service.Name(), service); err != nil {
			panic(err)
		}
	}

	return core
}

func (c *CoreService) defaultServices() ([]coreCommon.Service, error) {

	serviceList := []coreCommon.Service{

		builderapi.NewBuilderApiService(),
		blockaggregator.NewBlockAggregatorService(),
		relay.NewRelayService(),
	}

	return serviceList, nil
}

func (c *CoreService) Configure(coreConfig config.CoreConfig) error {

	for _, module := range c.moduleRegistry.Modules() {

		if _, ok := coreConfig.ModuleFlags[module.Name]; !ok {
			continue
		}

		if err := module.Service.Configure(coreConfig.ModuleFlags[module.Name]); err != nil {
			return fmt.Errorf("failed to configure module %s: %v", module.Name, err)
		}
	}

	knownCallbacks := make(map[string]bool)
	for _, module := range c.moduleRegistry.Modules() {
		for method := range module.Callbacks {
			knownCallbacks[module.Name+"_"+method] = true
		}
	}
	// Add ping callback
	knownCallbacks["core_ping"] = true

	coreClientContext := context.Background()
	_, coreClient, coreClientChannels, err := coreCommon.NewClient(coreClientContext, "core", nil, knownCallbacks)
	if err != nil {
		return fmt.Errorf("failed to create core client: %v", err)
	}
	c.coreClient = coreClient
	c.coreClientChannels = coreCommon.ModuleCommChannels{
		Incoming: coreClientChannels.Incoming,
		Outgoing: coreClientChannels.Outgoing,
	}

	log.Info("Set up Core Communication Client")

	for _, module := range c.moduleRegistry.Modules() {

		if _, ok := c.moduleChannels[module.Name]; ok {
			return fmt.Errorf("module %s already has a channel", module.Name)
		}

		moduleCtx := context.Background()
		moduleClientId, moduleClient, clientChans, err := coreCommon.NewClient(moduleCtx, module.Name, module.Callbacks, knownCallbacks)
		if err != nil {
			return fmt.Errorf("failed to create client for module %s: %v", module.Name, err)
		}
		incomingChan := clientChans.Incoming
		outgoingChan := clientChans.Outgoing

		// Connect the module to the core, and verify the connection
		// authenticate connection using a ping message
		pingId := c.idgen()
		pingMsg := fmt.Sprintf("%s_%s", moduleClientId, pingId)
		if err := module.Service.ConnectCore(moduleClient, pingMsg); err != nil {
			return fmt.Errorf("failed to connect module %s to core: %v", module.Name, err)
		}

		resp := <-outgoingChan

		if resp.Error != nil {
			return fmt.Errorf("failed to connect module %s to core: %v", module.Name, resp.Error)
		}
		var result []string
		if err := json.Unmarshal(resp.Params, &result); err != nil {
			return fmt.Errorf("failed to connect module %s to core: %v", module.Name, err)
		}
		if result[0] != pingMsg {
			return fmt.Errorf("failed to connect module %s to core: ping message mismatch", module.Name)
		}

		// Since these are private to the core, can be set directly.
		c.lock.Lock()
		c.moduleChannels[module.Name] = coreCommon.ModuleCommChannels{
			Incoming: incomingChan,
			Outgoing: outgoingChan,
		}
		// the module client ID is never exposed publicly to any other module
		// and is only known by the core and the module's client itself.
		c.moduleClientIds[module.Name] = moduleClientId
		// would allow the core to be the only entity to send a close message to the module

		c.lock.Unlock()

		log.Info("Connected module to core communication: ", module.Name)
	}

	return nil

}

func (c *CoreService) Start() error {

	c.startStopLock.Lock()
	defer c.startStopLock.Unlock()

	c.lock.Lock()
	switch c.state {
	case runningState:
		c.lock.Unlock()
		return common.ErrCoreRunning
	case closedState:
		c.lock.Unlock()
		return common.ErrCoreStopped
	}
	c.lock.Unlock()

	c.RelayComms()

	// Start all modules, that have been registered with the core service
	// as microservices if they are.
	_, err := c.moduleRegistry.StartModuleServices()
	if err != nil {
		// Stop all started modules
		_, stopErr := c.moduleRegistry.StopModuleServices()
		if stopErr != nil {
			err = fmt.Errorf("failed to stop modules after start error: %v, %v", err, stopErr)
			return stopErr
		}
		return err
	}

	c.lock.Lock()
	c.state = runningState
	c.lock.Unlock()

	return nil
}

func (c *CoreService) Close() error {
	c.startStopLock.Lock()
	defer c.startStopLock.Unlock()

	c.lock.Lock()
	state := c.state
	c.lock.Unlock()

	switch state {
	case initializingState:
		// The core service was not started,
		// however clients may have been created and connected and need to be closed.

		if err := c.close(); err != nil {
			return err
		}

		return nil
	case runningState:

		if err := c.close(); err != nil {
			return err
		}

		return nil

	case closedState:
		return common.ErrCoreStopped
	default:
		panic(fmt.Sprintf("core is in unknown state %d", state))
	}
}

func (c *CoreService) close() error {

	// After all module clients have been closed, close the core client
	if c.coreClient != nil {
		c.coreClient.Close()
	}

	// NB do not close any channels from the core the
	// clients would close their own channels when done

	_, err := c.moduleRegistry.StopModuleServices()
	if err != nil {
		return err
	}

	close(c.stop)

	c.lock.Lock()
	c.state = closedState
	c.lock.Unlock()

	return nil

}

// Wait blocks until the core is closed.
func (c *CoreService) Wait() {
	<-c.stop
}

// Config returns the configuration of the core
func (c *CoreService) Config() *config.CoreConfig {
	return &c.config
}

func (c *CoreService) RelayComms() {
	// Listen for incomming messages from all modules, and relay them to all other modules
	allModules := c.moduleRegistry.ModuleNames()
	for _, module := range allModules {

		// dont need a lock to access the channels as they would not be changed after they have been set
		if channels, ok := c.moduleChannels[module]; !ok {
			continue
		} else {
			go func(module string, channels coreCommon.ModuleCommChannels) {
				for {
					select {
					case <-c.stop:
						// All communication relays would be stopped when core is closed
						return
					case msg := <-channels.Outgoing:

						var targettedModule string
						if msg.IsResponse() {
							targettedModule = msg.Origin
						} else {
							targettedModule = msg.Namespace()
							if msg.Origin == "" {
								msg.Origin = module
							}
						}

						if targettedModule == module {
							// Targetted module is the same as the module that sent the message
							// Send to self in the case of a reverse call for instance
							channels.Incoming <- msg
						}

						targettedModuleChannels, ok := c.moduleChannels[targettedModule]
						if !ok {
							// Targetted module not found, send back to the module that sent the message if its not a notification
							if msg.IsCall() {

								errResponse := coreCommon.JsonRPCMessage{
									Version:   msg.Version,
									ID:        msg.ID,
									Method:    msg.Method + common.ResponseMethodSuffix,
									NotifyAll: false,
								}

								errResponse = *errResponse.ErrorResponse(fmt.Errorf("targetted module [%s] not found", targettedModule))

								channels.Incoming <- errResponse
							}

						} else {
							targettedModuleChannels.Incoming <- msg
						}

						if msg.NotifyAll {
							msgCopy := coreCommon.JsonRPCMessage{
								Version:   msg.Version,
								ID:        nil, // notifications cannot have an ID
								Method:    msg.Method,
								Params:    msg.Params,
								Error:     msg.Error,
								Result:    msg.Result,
								NotifyAll: msg.NotifyAll,
							}
							// if notify all, then the message sent to the rest of the modules
							// cannot require a response

							for _, otherModule := range allModules {
								if otherModule == module || otherModule == targettedModule {
									continue
								}
								// check if the module is in the notify exclusion list
								if msg.NotifyExclusion != nil {
									for _, exclusion := range msg.NotifyExclusion {
										if exclusion == otherModule {
											continue
										}
									}
								}

								otherModuleChannels, ok := c.moduleChannels[otherModule]
								if !ok {
									continue
								}
								otherModuleChannels.Incoming <- msgCopy
							}

						}

					}
				}
			}(module, channels)
		}
	}

}
