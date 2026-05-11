package cli

import (
	"fmt"
	"strings"

	sdkclob "github.com/TrebuchetDynamics/polygolem/pkg/clob"
	sdkdata "github.com/TrebuchetDynamics/polygolem/pkg/data"
	sdkorderresults "github.com/TrebuchetDynamics/polygolem/pkg/orderresults"
	"github.com/spf13/cobra"
)

func dataCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	cmd := commandGroup("data", "Polymarket Data API analytics")

	var user string
	var tokenID string
	var limit int

	addUser := func(c *cobra.Command) {
		c.Flags().StringVar(&user, "user", "", "user wallet address")
	}
	addUserLimit := func(c *cobra.Command) {
		addUser(c)
		c.Flags().IntVar(&limit, "limit", 20, "max rows")
	}
	requireUser := func() error {
		if strings.TrimSpace(user) == "" {
			return fmt.Errorf("--user required")
		}
		return nil
	}
	addToken := func(c *cobra.Command) {
		c.Flags().StringVar(&tokenID, "token-id", "", "CLOB token ID")
	}
	addTokenLimit := func(c *cobra.Command) {
		addToken(c)
		c.Flags().IntVar(&limit, "limit", 20, "max rows")
	}
	requireToken := func() error {
		if strings.TrimSpace(tokenID) == "" {
			return fmt.Errorf("--token-id required")
		}
		return nil
	}

	positionsCmd := &cobra.Command{Use: "positions", Short: "List open positions for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			rows, err := w.data.CurrentPositions(cmd.Context(), user)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addUserLimit(positionsCmd)
	cmd.AddCommand(positionsCmd)

	closedPositionsCmd := &cobra.Command{Use: "closed-positions", Short: "List closed positions for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			rows, err := w.data.ClosedPositions(cmd.Context(), user)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addUserLimit(closedPositionsCmd)
	cmd.AddCommand(closedPositionsCmd)

	tradesCmd := &cobra.Command{Use: "trades", Short: "List public Data API trades for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			rows, err := w.data.Trades(cmd.Context(), user, limit)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addUserLimit(tradesCmd)
	cmd.AddCommand(tradesCmd)

	var includeCLOB bool
	orderResultsCmd := &cobra.Command{Use: "order-results", Short: "Join positions, trades, and results for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			source := sdkorderresults.Source{
				Data: sdkdata.NewClient(sdkdata.Config{BaseURL: dataBaseURL}),
			}
			opts := sdkorderresults.Options{Limit: limit}
			if includeCLOB {
				key, err := privateKeyFromEnv()
				if err != nil {
					return err
				}
				opts.IncludeCLOB = true
				opts.PrivateKey = key
				cfg := sdkclob.Config{BaseURL: clobBaseURL}
				if creds, ok := clobL2CredentialsFromEnv(); ok {
					cfg.Credentials = sdkclob.APIKey{
						Key:        creds.Key,
						Secret:     creds.Secret,
						Passphrase: creds.Passphrase,
					}
				}
				source.CLOB = sdkclob.NewClient(cfg)
			}
			report, err := sdkorderresults.BuildReport(cmd.Context(), source, user, opts)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, report)
		},
	}
	addUserLimit(orderResultsCmd)
	orderResultsCmd.Flags().BoolVar(&includeCLOB, "include-clob", false, "include authenticated CLOB open orders and trade history")
	cmd.AddCommand(orderResultsCmd)

	activityCmd := &cobra.Command{Use: "activity", Short: "List public activity for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			rows, err := w.data.Activity(cmd.Context(), user, limit)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addUserLimit(activityCmd)
	cmd.AddCommand(activityCmd)

	holdersCmd := &cobra.Command{Use: "holders", Short: "List top holders for a token", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(); err != nil {
				return err
			}
			rows, err := w.data.TopHolders(cmd.Context(), tokenID, limit)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	addTokenLimit(holdersCmd)
	cmd.AddCommand(holdersCmd)

	valueCmd := &cobra.Command{Use: "value", Short: "Get total portfolio value for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			value, err := w.data.TotalValue(cmd.Context(), user)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, value)
		},
	}
	addUser(valueCmd)
	cmd.AddCommand(valueCmd)

	marketsTradedCmd := &cobra.Command{Use: "markets-traded", Short: "Get total markets traded for a user", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireUser(); err != nil {
				return err
			}
			value, err := w.data.MarketsTraded(cmd.Context(), user)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, value)
		},
	}
	addUser(marketsTradedCmd)
	cmd.AddCommand(marketsTradedCmd)

	openInterestCmd := &cobra.Command{Use: "open-interest", Short: "Get open interest for a token", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(); err != nil {
				return err
			}
			value, err := w.data.OpenInterest(cmd.Context(), tokenID)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, value)
		},
	}
	addToken(openInterestCmd)
	cmd.AddCommand(openInterestCmd)

	leaderboardCmd := &cobra.Command{Use: "leaderboard", Short: "List trader leaderboard rows", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rows, err := w.data.TraderLeaderboard(cmd.Context(), limit)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, rows)
		},
	}
	leaderboardCmd.Flags().IntVar(&limit, "limit", 20, "max rows")
	cmd.AddCommand(leaderboardCmd)

	liveVolumeCmd := &cobra.Command{Use: "live-volume", Short: "Get live volume summary", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := w.data.LiveVolume(cmd.Context(), limit)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, value)
		},
	}
	liveVolumeCmd.Flags().IntVar(&limit, "limit", 20, "max rows")
	cmd.AddCommand(liveVolumeCmd)

	return cmd
}
