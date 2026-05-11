package cli

import (
	"fmt"
	"strings"

	"github.com/TrebuchetDynamics/polygolem/internal/polytypes"
	"github.com/spf13/cobra"
)

func discoverCmd(jsonOut bool) *cobra.Command {
	w := newWire(jsonOut)
	var query, marketID, marketSlug string
	var limit int

	cmd := commandGroup("discover", "Market discovery via Polymarket Gamma API")

	var marketsLimit, marketsOffset, marketsTagID int
	var marketsOrder string
	var marketsActive, marketsClosed, marketsAscending bool
	marketsCmd := &cobra.Command{Use: "markets", Short: "List Gamma markets", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params := &polytypes.GetMarketsParams{
				Limit:     marketsLimit,
				Offset:    marketsOffset,
				Order:     marketsOrder,
				Active:    &marketsActive,
				Closed:    &marketsClosed,
				Ascending: &marketsAscending,
			}
			if marketsTagID > 0 {
				params.TagID = &marketsTagID
			}
			markets, err := w.gamma.Markets(cmd.Context(), params)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, markets)
		},
	}
	marketsCmd.Flags().IntVar(&marketsLimit, "limit", 20, "max markets")
	marketsCmd.Flags().IntVar(&marketsOffset, "offset", 0, "pagination offset")
	marketsCmd.Flags().StringVar(&marketsOrder, "order", "", "Gamma order field")
	marketsCmd.Flags().BoolVar(&marketsActive, "active", true, "filter active markets")
	marketsCmd.Flags().BoolVar(&marketsClosed, "closed", false, "filter closed markets")
	marketsCmd.Flags().BoolVar(&marketsAscending, "ascending", false, "sort ascending")
	marketsCmd.Flags().IntVar(&marketsTagID, "tag-id", 0, "filter by tag id")
	cmd.AddCommand(marketsCmd)

	searchCmd := &cobra.Command{
		Use: "search", Short: "Search markets and events", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if query != "" {
				resp, err := w.gamma.Search(cmd.Context(), &polytypes.SearchParams{Q: query, LimitPerType: &limit})
				if err != nil {
					return err
				}
				return w.printJSON(cmd, resp)
			}
			active, closed := true, false
			markets, err := w.gamma.Markets(cmd.Context(), &polytypes.GetMarketsParams{Active: &active, Closed: &closed, Limit: limit})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, markets)
		},
	}
	searchCmd.Flags().StringVar(&query, "query", "", "text query")
	searchCmd.Flags().IntVar(&limit, "limit", 10, "max results")
	cmd.AddCommand(searchCmd)

	marketCmd := &cobra.Command{
		Use: "market", Short: "Get market details", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if marketID != "" {
				m, err := w.gamma.MarketByID(cmd.Context(), marketID)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, m)
			}
			if marketSlug != "" {
				m, err := w.gamma.MarketBySlug(cmd.Context(), marketSlug)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, m)
			}
			return fmt.Errorf("--id or --slug required")
		},
	}
	marketCmd.Flags().StringVar(&marketID, "id", "", "market Gamma ID")
	marketCmd.Flags().StringVar(&marketSlug, "slug", "", "market slug")
	cmd.AddCommand(marketCmd)

	enrichCmd := &cobra.Command{
		Use: "enrich", Short: "Enrich market with CLOB data", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if marketID == "" {
				return fmt.Errorf("--id required")
			}
			m, err := w.gamma.MarketByID(cmd.Context(), marketID)
			if err != nil {
				return err
			}
			em, err := w.discover.EnrichMarket(cmd.Context(), *m)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, em)
		},
	}
	enrichCmd.Flags().StringVar(&marketID, "id", "", "market Gamma ID")
	cmd.AddCommand(enrichCmd)

	var tagsLimit, tagsOffset int
	var tagID, tagSlug string
	tagsCmd := &cobra.Command{Use: "tags", Short: "List or fetch Gamma tags/categories", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tagID != "" {
				tag, err := w.gamma.TagByID(cmd.Context(), tagID)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, tag)
			}
			if tagSlug != "" {
				tag, err := w.gamma.TagBySlug(cmd.Context(), tagSlug)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, tag)
			}
			tags, err := w.gamma.Tags(cmd.Context(), &polytypes.GetTagsParams{
				Limit:  tagsLimit,
				Offset: tagsOffset,
			})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, tags)
		},
	}
	tagsCmd.Flags().StringVar(&tagID, "id", "", "tag ID")
	tagsCmd.Flags().StringVar(&tagSlug, "slug", "", "tag slug")
	tagsCmd.Flags().IntVar(&tagsLimit, "limit", 100, "max tags")
	tagsCmd.Flags().IntVar(&tagsOffset, "offset", 0, "pagination offset")
	cmd.AddCommand(tagsCmd)

	var seriesLimit, seriesOffset int
	var seriesID string
	var seriesClosed bool
	seriesCmd := &cobra.Command{Use: "series", Short: "List or fetch Gamma series", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if seriesID != "" {
				series, err := w.gamma.SeriesByID(cmd.Context(), seriesID)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, series)
			}
			series, err := w.gamma.Series(cmd.Context(), &polytypes.GetSeriesParams{
				Limit:  seriesLimit,
				Offset: seriesOffset,
				Closed: &seriesClosed,
			})
			if err != nil {
				return err
			}
			return w.printJSON(cmd, series)
		},
	}
	seriesCmd.Flags().StringVar(&seriesID, "id", "", "series ID")
	seriesCmd.Flags().IntVar(&seriesLimit, "limit", 20, "max series")
	seriesCmd.Flags().IntVar(&seriesOffset, "offset", 0, "pagination offset")
	seriesCmd.Flags().BoolVar(&seriesClosed, "closed", false, "filter closed series")
	cmd.AddCommand(seriesCmd)

	var commentID, commentEntityType, commentUser string
	var commentEntityID, commentLimit, commentOffset int
	commentsCmd := &cobra.Command{Use: "comments", Short: "List or fetch public Gamma comments", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if commentID != "" {
				comment, err := w.gamma.CommentByID(cmd.Context(), commentID)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, comment)
			}
			if commentUser != "" {
				comments, err := w.gamma.CommentsByUser(cmd.Context(), commentUser, commentLimit)
				if err != nil {
					return err
				}
				return w.printJSON(cmd, comments)
			}
			params := &polytypes.CommentQuery{Limit: commentLimit, Offset: commentOffset}
			if commentEntityID > 0 {
				params.EntityID = &commentEntityID
			}
			if commentEntityType != "" {
				params.EntityType = &commentEntityType
			}
			comments, err := w.gamma.Comments(cmd.Context(), params)
			if err != nil {
				return err
			}
			return w.printJSON(cmd, comments)
		},
	}
	commentsCmd.Flags().StringVar(&commentID, "id", "", "comment ID")
	commentsCmd.Flags().StringVar(&commentUser, "user", "", "user wallet address")
	commentsCmd.Flags().IntVar(&commentEntityID, "entity-id", 0, "comment parent entity ID")
	commentsCmd.Flags().StringVar(&commentEntityType, "entity-type", "", "comment parent entity type")
	commentsCmd.Flags().IntVar(&commentLimit, "limit", 20, "max comments")
	commentsCmd.Flags().IntVar(&commentOffset, "offset", 0, "pagination offset")
	cmd.AddCommand(commentsCmd)

	var cryptoInterval string
	var cryptoAsset string
	var cryptoEnrich bool
	cryptoCmd := &cobra.Command{
		Use:   "crypto",
		Short: "Discover active crypto prediction markets",
		Long: `Search for active Polymarket crypto markets by asset and interval.

Extracts markets from events and filters by title patterns. Returns token IDs
ready for orderbook inspection or trading.

Examples:
  polygolem discover crypto --asset BTC --interval 5m    # BTC Up/Down 5m markets
  polygolem discover crypto --asset ETH --interval 15m   # ETH Up/Down 15m markets
  polygolem discover crypto --asset BTC --interval 5m --enrich  # With CLOB prices
  polygolem discover crypto --limit 50                   # All crypto markets

Assets: BTC, ETH, SOL, XRP, DOGE, BNB, HYPE, etc.
Intervals: 5m, 15m, 1h, daily, weekly (matches title patterns)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			searchQuery := cryptoAsset
			if cryptoInterval != "" {
				if searchQuery != "" {
					searchQuery += " "
				}
				searchQuery += cryptoInterval
			}
			if searchQuery == "" {
				searchQuery = "crypto"
			}
			resp, err := w.gamma.Search(cmd.Context(), &polytypes.SearchParams{
				Q:            searchQuery,
				LimitPerType: &limit,
			})
			if err != nil {
				return err
			}

			// Extract markets from events and filter
			type cryptoMarket struct {
				EventID      string   `json:"event_id"`
				EventTitle   string   `json:"event_title"`
				EventSlug    string   `json:"event_slug"`
				MarketID     string   `json:"market_id"`
				Question     string   `json:"question"`
				ConditionID  string   `json:"condition_id"`
				TokenID      string   `json:"token_id"`
				Outcomes     []string `json:"outcomes"`
				OutcomePrices []string `json:"outcome_prices"`
				EndDate      string   `json:"end_date"`
				Volume24hr   float64  `json:"volume_24h"`
				Price        string   `json:"price,omitempty"`
				Spread       string   `json:"spread,omitempty"`
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
					// Filter by asset if specified
					if cryptoAsset != "" && !strings.Contains(strings.ToUpper(market.Question), strings.ToUpper(cryptoAsset)) &&
						!strings.Contains(strings.ToUpper(event.Title), strings.ToUpper(cryptoAsset)) {
						continue
					}
					// Filter by interval if specified
					if cryptoInterval != "" && !strings.Contains(strings.ToLower(event.Title), strings.ToLower(cryptoInterval)) &&
						!strings.Contains(strings.ToLower(market.Question), strings.ToLower(cryptoInterval)) {
						continue
					}

					cm := cryptoMarket{
						EventID:     event.ID,
						EventTitle:  event.Title,
						EventSlug:   event.Slug,
						MarketID:    market.ID,
						Question:    market.Question,
						ConditionID: market.ConditionID,
						TokenID:     market.ClobTokenIDs,
						EndDate:     market.EndDateISO,
						Volume24hr:  market.Volume24hr,
					}
					cm.Outcomes = []string(market.Outcomes)
					cm.OutcomePrices = []string(market.OutcomePrices)

					if cryptoEnrich {
						tokenIDs := parseClobTokenIDs(market.ClobTokenIDs)
						if len(tokenIDs) > 0 {
							if price, err := w.clob.Price(cmd.Context(), tokenIDs[0], "BUY"); err == nil {
								cm.Price = price
							}
							if spread, err := w.clob.Spread(cmd.Context(), tokenIDs[0]); err == nil {
								cm.Spread = spread
							}
						}
					}

					results = append(results, cm)
				}
			}

			return w.printJSON(cmd, map[string]interface{}{
				"query":       searchQuery,
				"asset":       cryptoAsset,
				"interval":    cryptoInterval,
				"count":       len(results),
				"markets":     results,
			})
		},
	}
	cryptoCmd.Flags().StringVar(&cryptoAsset, "asset", "", "crypto asset filter (BTC, ETH, SOL, XRP, DOGE, BNB, HYPE)")
	cryptoCmd.Flags().StringVar(&cryptoInterval, "interval", "", "interval filter (5m, 15m, 1h, daily, weekly)")
	cryptoCmd.Flags().IntVar(&limit, "limit", 20, "max results")
	cryptoCmd.Flags().BoolVar(&cryptoEnrich, "enrich", false, "enrich with CLOB price and spread (slower, one API call per market)")
	cmd.AddCommand(cryptoCmd)

	return cmd
}
