package main

import (
	"os"

	"github.com/TrebuchetDynamics/polymarket-cli-go/internal/cli"
)

var version = "dev"

func main() {
	root := cli.NewRootCommand(cli.Options{Version: version})
	if err := root.Execute(); err != nil {
		_, _ = root.ErrOrStderr().Write([]byte(err.Error() + "\n"))
		os.Exit(1)
	}
}
