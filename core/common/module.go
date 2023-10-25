package common

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pon-pbs/mev-plus/common"
)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	stringType  = reflect.TypeOf("")
)

type ModuleRegistry struct {
	mu      sync.Mutex
	modules map[string]Module
}

func (r *ModuleRegistry) RegisterName(name string, rcvr Service) error {

	name = strings.TrimSpace(formatName(name))
	rcvrVal := reflect.ValueOf(rcvr)
	if name == "" {
		return fmt.Errorf("no service name for type %s", rcvrVal.Type().String())
	}
	callbacks := suitableCallbacks(rcvrVal)
	if len(callbacks) == 0 {
		return fmt.Errorf("service %T doesn't have any suitable methods to expose", rcvr)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.modules == nil {
		r.modules = make(map[string]Module)
	}
	module, ok := r.modules[name]
	if !ok {
		module = Module{
			Name:      rcvr.Name(),
			Service:   rcvr,
			Callbacks: make(map[string]*Callback),
		}
		r.modules[name] = module
	}
	for name, cb := range callbacks {
		module.Callbacks[name] = cb
	}
	return nil
}

func (r *ModuleRegistry) StartModuleServices() (started []string, err error) {

	knownModules := r.ModuleNames()
	started = make([]string, 0, len(knownModules))

	for _, moduleName := range knownModules {
		if err := r.startModuleService(moduleName); err != nil {
			return started, err
		}
		started = append(started, moduleName)
	}
	return started, err
}

func (r *ModuleRegistry) startModuleService(moduleName string) error {
	done := make(chan error) // Channel to signal early return with error

	r.mu.Lock()
	defer r.mu.Unlock()
	defer close(done)

	module, ok := r.modules[moduleName]
	if !ok {
		return fmt.Errorf("module %s not found", moduleName)
	}

	defer func() {
		r.modules[moduleName] = module
	}()

	startTimer := time.NewTimer(30 * time.Second)
	go func() {
		if err := module.Service.Start(); err != nil {
			done <- fmt.Errorf("failed to start module %s: %v", module.Name, err)
			return
		}
		startTimer.Stop()
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			module.ServiceAlive = false
		} else {
			module.ServiceAlive = true
		}
		return err
	case <-startTimer.C:
		module.ServiceAlive = false
		return fmt.Errorf("module %s start took too long or may be blocking", module.Name)
	}
}

func (r *ModuleRegistry) StopModuleServices() (stopped []string, err error) {

	knownModules := r.ModuleNames()
	stopped = make([]string, 0, len(knownModules))
	failures := make(map[string]error)

	for _, moduleName := range knownModules {
		r.mu.Lock()
		if r.modules[moduleName].ServiceAlive == false {
			r.mu.Unlock()
			continue
		}
		r.mu.Unlock()
		err := r.stopModuleService(moduleName)
		if err != nil {
			failures[moduleName] = err
		} else {
			stopped = append(stopped, moduleName)
		}
	}

	if len(failures) > 0 {
		err = fmt.Errorf("failed to stop modules: %v", failures)
	}

	return stopped, err
}

func (r *ModuleRegistry) stopModuleService(moduleName string) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	module, ok := r.modules[moduleName]
	if !ok {
		return fmt.Errorf("module %s not found", moduleName)
	}

	defer func() {
		r.modules[moduleName] = module
	}()

	if module.ServiceAlive == false {
		return fmt.Errorf("module %s is not alive", moduleName)
	}

	err := module.Service.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop module %s: %v", module.Name, err)
	} else {
		module.ServiceAlive = false
	}

	return nil

}

func (r *ModuleRegistry) Modules() []Module {
	r.mu.Lock()
	defer r.mu.Unlock()
	var modules []Module
	for _, service := range r.modules {
		modules = append(modules, service)
	}
	return modules
}

func (r *ModuleRegistry) ModuleNames() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var names []string
	for name := range r.modules {
		names = append(names, name)
	}
	return names
}

// callback returns the callback corresponding to the given method name.
func (r *ModuleRegistry) callback(method string) *Callback {
	elem := strings.SplitN(method, common.ServiceMethodSeparator, 2)
	if len(elem) != 2 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return r.modules[elem[0]].Callbacks[elem[1]]
}
