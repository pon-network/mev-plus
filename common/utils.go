package common

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"strings"
	"sync"
	"time"

	"fmt"

	cli "github.com/urfave/cli/v2"
)

var globalIDGen = randomIDGenerator()

// NewID returns a new, random ID.
func NewID() string {
	return globalIDGen()
}

// randomIDGenerator returns a function generates a random IDs.
func randomIDGenerator() func() string {
	var buf = make([]byte, 8)
	var seed int64
	if _, err := crand.Read(buf); err == nil {
		seed = int64(binary.BigEndian.Uint64(buf))
	} else {
		seed = int64(time.Now().Nanosecond())
	}

	var (
		mu  sync.Mutex
		rng = rand.New(rand.NewSource(seed))
	)
	return func() string {
		mu.Lock()
		defer mu.Unlock()
		id := make([]byte, 16)
		rng.Read(id)
		return encodeID(id)
	}
}

func encodeID(b []byte) string {
	id := hex.EncodeToString(b)
	id = strings.TrimLeft(id, "0")
	if id == "" {
		id = "0"
	}
	return "0x" + id
}


func FormatCommands(commands []*cli.Command) ([]*cli.Command, error) {
	var moduleCommandList []*cli.Command
	var moduleNameMap = make(map[string]bool)

	// Load default modules
	for _, m := range DefaultModuleNames {
		moduleNameMap[m] = true
	}

	for _, command := range commands {
		name, err := FormatToAllowed(command.Name)
		if err != nil {
			return nil, err
		}
		command.Name = name
		command.Category = strings.ToUpper(strings.TrimSpace(command.Category))

		if _, ok := moduleNameMap[command.Name]; ok {
			return nil, fmt.Errorf("Duplicate module command name %s", command.Name)
		}
		moduleNameMap[command.Name] = true
		moduleCommandList = append(moduleCommandList, command)
	}
	return moduleCommandList, nil
}

func FormatToAllowed(name string) (formattedName string, err error) {

	// ensure names do not have space/underscore in it, if it does raise error
	if notOk := strings.ContainsAny(name, " _-%"); notOk {
		return name, fmt.Errorf("module name (%s) must not contain underscores/spaces", name)
	}

	formattedName = strings.ToLower(strings.TrimSpace(name))

	return formattedName, nil

}
