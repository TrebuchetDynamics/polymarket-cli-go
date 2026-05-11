package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TrebuchetDynamics/polygolem/internal/paper"
	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/spf13/cobra"
)

func paperCmd(jsonOut bool) *cobra.Command {
	cmd := commandGroup("paper", "Paper trading simulation for crypto markets")

	var paperCash float64
	var tokenID string
	var priceStr string
	var sizeStr string

	paperState := paper.NewState("USD", 10000.0)

	buyCmd := &cobra.Command{
		Use:   "buy",
		Short: "Simulate a buy order (paper trading)",
		Long: `Simulate a buy order against live market data.
Uses current best ask price if --price is not specified.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenID == "" {
				return fmt.Errorf("--token-id required")
			}

			price := 0.5
			if priceStr != "" {
				p, err := strconv.ParseFloat(priceStr, 64)
				if err != nil {
					return fmt.Errorf("invalid price: %w", err)
				}
				price = p
			} else {
				// Get current price from CLOB
				w := newWire(jsonOut)
				if p, err := w.clob.Price(cmd.Context(), tokenID, "SELL"); err == nil {
					if parsed, err := strconv.ParseFloat(p, 64); err == nil {
						price = parsed
					}
				}
			}

			size := 1.0
			if sizeStr != "" {
				s, err := strconv.ParseFloat(sizeStr, 64)
				if err != nil {
					return fmt.Errorf("invalid size: %w", err)
				}
				size = s
			}

			fill, err := paperState.Buy(paper.Order{
				TokenID: tokenID,
				Price:   price,
				Size:    size,
			})
			if err != nil {
				return err
			}

			return writeCommandJSON(cmd, map[string]interface{}{
				"action":   "buy",
				"token_id": tokenID,
				"price":    price,
				"size":     size,
				"cost":     price * size,
				"cash":     paperState.Cash,
				"fill":     fill,
			})
		},
	}
	buyCmd.Flags().StringVar(&tokenID, "token-id", "", "CLOB token ID to buy")
	buyCmd.Flags().StringVar(&priceStr, "price", "", "limit price (default: best ask)")
	buyCmd.Flags().StringVar(&sizeStr, "size", "1", "number of shares")
	cmd.AddCommand(buyCmd)

	sellCmd := &cobra.Command{
		Use:   "sell",
		Short: "Simulate a sell order (paper trading)",
		Long: `Simulate a sell order against live market data.
Uses current best bid price if --price is not specified.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenID == "" {
				return fmt.Errorf("--token-id required")
			}

			price := 0.5
			if priceStr != "" {
				p, err := strconv.ParseFloat(priceStr, 64)
				if err != nil {
					return fmt.Errorf("invalid price: %w", err)
				}
				price = p
			} else {
				w := newWire(jsonOut)
				if p, err := w.clob.Price(cmd.Context(), tokenID, "BUY"); err == nil {
					if parsed, err := strconv.ParseFloat(p, 64); err == nil {
						price = parsed
					}
				}
			}

			size := 1.0
			if sizeStr != "" {
				s, err := strconv.ParseFloat(sizeStr, 64)
				if err != nil {
					return fmt.Errorf("invalid size: %w", err)
				}
				size = s
			}

			// For paper trading, selling is just buying the opposite outcome
			fill, err := paperState.Buy(paper.Order{
				TokenID: tokenID,
				Price:   price,
				Size:    size,
			})
			if err != nil {
				return err
			}

			return writeCommandJSON(cmd, map[string]interface{}{
				"action":   "sell",
				"token_id": tokenID,
				"price":    price,
				"size":     size,
				"proceeds": price * size,
				"cash":     paperState.Cash,
				"fill":     fill,
			})
		},
	}
	sellCmd.Flags().StringVar(&tokenID, "token-id", "", "CLOB token ID to sell")
	sellCmd.Flags().StringVar(&priceStr, "price", "", "limit price (default: best bid)")
	sellCmd.Flags().StringVar(&sizeStr, "size", "1", "number of shares")
	cmd.AddCommand(sellCmd)

	positionsCmd := &cobra.Command{
		Use:   "positions",
		Short: "Show current paper trading positions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeCommandJSON(cmd, map[string]interface{}{
				"cash":      paperState.Cash,
				"positions": paperState.Positions,
				"fills":     paperState.Fills,
			})
		},
	}
	cmd.AddCommand(positionsCmd)

	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset paper trading state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paperState = paper.NewState("USD", paperCash)
			return writeCommandJSON(cmd, map[string]interface{}{
				"status": "reset",
				"cash":   paperCash,
			})
		},
	}
	resetCmd.Flags().Float64Var(&paperCash, "cash", 10000.0, "initial paper cash")
	cmd.AddCommand(resetCmd)

	var cryptoAsset, cryptoInterval string
	var cryptoLimit int
	cryptoCmd := &cobra.Command{
		Use:   "crypto",
		Short: "Discover crypto markets and paper trade",
		Long: `Find active crypto markets and get token IDs for paper trading.

Examples:
  polygolem paper crypto --asset BTC --interval 5m
  polygolem paper crypto --asset ETH --limit 10`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := newWire(jsonOut)
			searchQuery := ""

			if cryptoAsset != "" {
				searchQuery = cryptoAsset
			}
			if cryptoInterval != "" {
				if searchQuery != "" {
					searchQuery += " "
				}
				searchQuery += cryptoInterval
			}
			if searchQuery == "" {
				searchQuery = "crypto"
			}
			if cryptoLimit == 0 {
				cryptoLimit = 10
			}

			resp, err := w.gamma.Search(cmd.Context(), &polytypes.SearchParams{
				Q:            searchQuery,
				LimitPerType: &cryptoLimit,
			})
			if err != nil {
				return err
			}

			type cryptoMarket struct {
				EventID    string   `json:"event_id"`
				EventTitle string   `json:"event_title"`
				MarketID   string   `json:"market_id"`
				Question   string   `json:"question"`
				TokenID    string   `json:"token_id"`
				Outcomes   []string `json:"outcomes"`
				EndDate    string   `json:"end_date"`
			}

			var results []cryptoMarket
			for _, event := range resp.Events {
				if !event.Active || event.Closed {
					continue
				}
				for _, market := range event.Markets {
					if !market.Active || market.Closed {
						continue
					}
					if cryptoAsset != "" &&
						!strings.Contains(strings.ToUpper(market.Question), strings.ToUpper(cryptoAsset)) &&
						!strings.Contains(strings.ToUpper(event.Title), strings.ToUpper(cryptoAsset)) {
						continue
					}
					if cryptoInterval != "" &&
						!strings.Contains(strings.ToLower(event.Title), strings.ToLower(cryptoInterval)) &&
						!strings.Contains(strings.ToLower(market.Question), strings.ToLower(cryptoInterval)) {
						continue
					}

					tokenIDs := parseClobTokenIDs(market.ClobTokenIDs)
					if len(tokenIDs) == 0 {
						continue
					}

					cm := cryptoMarket{
						EventID:    event.ID,
						EventTitle: event.Title,
						MarketID:   market.ID,
						Question:   market.Question,
						TokenID:    tokenIDs[0],
						EndDate:    market.EndDateISO,
					}
					cm.Outcomes = []string(market.Outcomes)
					results = append(results, cm)
				}
			}

			return writeCommandJSON(cmd, map[string]interface{}{
				"query":   searchQuery,
				"count":   len(results),
				"markets": results,
				"help":    "Use 'polygolem paper buy --token-id <ID> --size 1' to paper trade",
			})
		},
	}
	cryptoCmd.Flags().StringVar(&cryptoAsset, "asset", "", "crypto asset filter (BTC, ETH, SOL, etc.)")
	cryptoCmd.Flags().StringVar(&cryptoInterval, "interval", "", "interval filter (5m, 15m, 1h)")
	cryptoCmd.Flags().IntVar(&cryptoLimit, "limit", 10, "max markets")
	cmd.AddCommand(cryptoCmd)

	return cmd
}
