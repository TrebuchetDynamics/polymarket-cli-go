package cli

import (
	"github.com/TrebuchetDynamics/polygolem/pkg/bridge"
	"github.com/spf13/cobra"
)

func bridgeCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	cmd := commandGroup("bridge", "Polymarket Bridge API")
	cmd.AddCommand(&cobra.Command{
		Use: "assets", Short: "List supported bridge assets", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			bc := bridge.NewClient("", nil)
			a, err := bc.GetSupportedAssets(cmd.Context())
			if err != nil {
				return err
			}
			return w.printJSON(cmd, a)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use: "deposit <address>", Short: "Create deposit addresses", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bc := bridge.NewClient("", nil)
			d, err := bc.CreateDepositAddress(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return w.printJSON(cmd, d)
		},
	})
	return cmd
}
