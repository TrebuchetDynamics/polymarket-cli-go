package cli

import (
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/spf13/cobra"
)

func eventsCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	var limit int
	cmd := commandGroup("events", "List Polymarket events")
	cmd.AddCommand(&cobra.Command{
		Use: "list", Short: "List events", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			events, err := w.gamma.Events(cmd.Context(), &polytypes.GetEventsParams{Limit: limit})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, events)
		},
	})
	return cmd
}
