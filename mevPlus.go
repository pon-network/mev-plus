package main

import (
	core "github.com/bsn-eng/mev-plus/cmd/core"
)

func main() {
	core.Run()
}

// In order to access the cli directly instead of go run mevPlus.go [command] [flags]
// build the binary with go build -o mevPlus mevPlus.go
// then run ./mevPlus [command] [flags]
