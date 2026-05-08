package main

import (
	"os"

	"github.com/TrebuchetDynamics/polygolem/internal/cli"
)

var version = "dev"

func main() {
	root := cli.NewRootCommand(cli.Options{Version: version})
	if err := root.Execute(); err != nil {
		if !cli.ErrorAlreadyRendered(err) {
			_, _ = root.ErrOrStderr().Write([]byte(err.Error() + "\n"))
		}
		os.Exit(cli.ExitCode(err))
	}
}
