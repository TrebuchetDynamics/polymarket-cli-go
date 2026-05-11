package cli

import "github.com/spf13/cobra"

func healthCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	return &cobra.Command{
		Use: "health", Short: "Check Gamma and CLOB API reachability", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			status := map[string]string{"gamma": "ok", "clob": "ok"}
			if _, err := w.gamma.HealthCheck(cmd.Context()); err != nil {
				status["gamma"] = err.Error()
			}
			if err := w.clob.Health(cmd.Context()); err != nil {
				status["clob"] = err.Error()
			}
			return w.printJSON(cmd, status)
		},
	}
}
