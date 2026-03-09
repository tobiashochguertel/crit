package main

import (
	"os"

	"github.com/tobiashochguertel/crit/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
