package main

import (
	"os"

	"github.com/tobiashochguertel/crit/internal/cli"
)

// version is injected at build time via goreleaser ldflags:
//
//	-X main.version={{.Version}}
var version = "dev"

func main() {
	os.Exit(cli.Execute(version))
}
