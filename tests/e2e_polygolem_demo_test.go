package tests

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/polygolem/pkg/types"
	"github.com/TrebuchetDynamics/polygolem/pkg/universal"
)

// demoMarket holds a diverse set of active markets for comprehensive testing.
// Categories: entertainment, politics, crypto, sports, science.
var demoMarkets = []struct {
	name        string
	id          string
	slug        string
	eventID     int
	category    string
	conditionID string
	tokens      []string
}{
	{
		name:     "Rihanna Album vs GTA VI",
		id:       "540817",
		slug:     "new-rhianna-album-before-gta-vi-926",
		eventID:  23784,
		category: "entertainment",
		conditionID: "0x1fad72fae204143ff1c3035e99e7c0f65ea8d5cd9bd1070987bd1a3316f772be",
		tokens: []string{
			"98022490269692409998126496127597032490334070080325855126491859374983463996227",
			"53831553061883006530739877284105938919721408776239639687877978808906551086026",
		},
	},
}

// TestPolygolemFullDemo is the canonical end-to-end demonstration of polygolem's
// public SDK capabilities against live Polymarket APIs.
//
// It extracts every available data point, validates consistency across
// endpoints, stresses the client concurrently, and logs a structured summary.
// Skipped under -short; requires network access.
func TestPolygolemFullDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live demo in short mode")
	}

	client := universal.NewClient(universal.Config{})
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	report := &demoReport{start: time.Now()}

	// Phase 1: Health & Discovery.
	t.Run("health_and_discovery", func(t *testing.T) {
		h, err := client.HealthCheck(ctx)
		if err != nil {
			t.Fatalf("HealthCheck: %v", err)
		}
		report.health = fmt.Sprintf("gamma=%v clob=%v data=%v", h.GammaOK, h.CLOBOK, h.DataOK)
		t.Logf("Health: %s", report.health)

		markets, err := client.Markets(ctx, &types.GetMarketsParams{Limit: 20, Active: boolPtr(true)})
		if err != nil {
			t.Fatalf("Markets: %v", err)
		}
		report.discoveredMarkets = len(markets)

		events, _ := client.Events(ctx, &types.GetEventsParams{Limit: 10})
		report.discoveredEvents = len(events)

		tags, _ := client.Tags(ctx, &types.GetTagsParams{Limit: 10})
		report.discoveredTags = len(tags)

		series, _ := client.Series(ctx, &types.GetSeriesParams{Limit: 10})
		report.discoveredSeries = len(series)

		t.Logf("Discovered: %d markets, %d events, %d tags, %d series",
			report.discoveredMarkets, report.discoveredEvents,
			report.discoveredTags, report.discoveredSeries)
	})

	// Phase 2: Deep extraction on primary market.
	t.Run("deep_extraction", func(t *testing.T) {
		m := demoMarkets[0]
		token := m.tokens[0]

		// Gamma metadata.
		market, err := client.MarketByID(ctx, m.id)
		if err != nil {
			t.Fatalf("MarketByID: %v", err)
		}
		report.marketName = market.Question
		report.marketPrices = fmt.Sprintf("%v", market.OutcomePrices)
		report.liquidity = market.LiquidityNum
		report.volume = market.VolumeNum

		// CLOB order book.
		book, err := client.OrderBook(ctx, token)
		if err != nil {
			t.Fatalf("OrderBook: %v", err)
		}
		report.bids = len(book.Bids)
		report.asks = len(book.Asks)

		// Price surface.
		buy, _ := client.Price(ctx, token, "buy")
		sell, _ := client.Price(ctx, token, "sell")
		mid, _ := client.Midpoint(ctx, token)
		spread, _ := client.Spread(ctx, token)
		last, _ := client.LastTradePrice(ctx, token)
		report.priceBuy = buy
		report.priceSell = sell
		report.midpoint = mid
		report.spread = spread
		report.lastTrade = last

		// Market microstructure.
		tick, _ := client.TickSize(ctx, token)
		fee, _ := client.FeeRateBps(ctx, token)
		nr, _ := client.NegRisk(ctx, token)
		report.tickSize = tick.MinimumTickSize
		report.feeBps = fee
		report.negRisk = nr.NegRisk

		// CLOB metadata.
		clobMkt, _ := client.CLOBMarket(ctx, m.conditionID)
		if clobMkt != nil {
			report.acceptingOrders = clobMkt.AcceptingOrders
		}

		clobByToken, _ := client.CLOBMarketByToken(ctx, token)
		if clobByToken != nil {
			report.conditionID = clobByToken.ConditionID
		}

		// Historical prices.
		endTS := time.Now().Unix()
		hist, _ := client.PricesHistory(ctx, &types.CLOBPriceHistoryParams{
			Market:   m.conditionID,
			StartTS:  endTS - 3600,
			EndTS:    endTS,
			Fidelity: 60,
		})
		if hist != nil {
			report.priceHistoryPoints = len(hist.History)
		}

		// Comments.
		entityType := "Event"
		comments, _ := client.Comments(ctx, &types.CommentQuery{
			EntityID:   &m.eventID,
			EntityType: &entityType,
			Limit:      5,
		})
		report.comments = len(comments)

		// Search.
		search, _ := client.Search(ctx, &types.SearchParams{
			Q:            "Rihanna",
			LimitPerType: intPtr(5),
		})
		report.searchResults = len(search.Events)

		// Simplified markets.
		simplified, _ := client.SimplifiedMarkets(ctx, "")
		if simplified != nil {
			report.simplifiedMarkets = len(simplified.Data)
		}

		// Server time.
		serverTime, _ := client.CLOBServerTime(ctx)
		if serverTime != nil {
			report.serverTime = serverTime.ISO
		}

		// Batch operations.
		books, _ := client.OrderBooks(ctx, []types.CLOBBookParams{{TokenID: token}})
		report.batchBooks = len(books)

		prices, _ := client.Prices(ctx, []types.CLOBBookParams{{TokenID: token, Side: "BUY"}})
		report.batchPrices = len(prices)

		// CLOB markets list.
		clobMarkets, _ := client.CLOBMarkets(ctx, "")
		if clobMarkets != nil {
			report.clobMarkets = len(clobMarkets.Data)
		}

		// Sampling markets.
		sampling, _ := client.SamplingMarkets(ctx, "")
		if sampling != nil {
			report.samplingMarkets = len(sampling.Data)
		}

		// Leaderboard.
		leaderboard, _ := client.TraderLeaderboard(ctx, 5)
		report.leaderboard = len(leaderboard)

		// Open interest.
		oi, _ := client.OpenInterest(ctx, m.conditionID)
		if oi != nil {
			report.openInterest = oi.OpenValue
		}

		// Top holders.
		holders, _ := client.TopHolders(ctx, m.conditionID, 3)
		report.topHolders = len(holders)

		// Live volume.
		lv, _ := client.LiveVolume(ctx, m.eventID)
		if lv != nil {
			report.liveVolume = lv.Total
		}

		// Event by ID.
		event, _ := client.EventByID(ctx, fmt.Sprintf("%d", m.eventID))
		if event != nil {
			report.eventTitle = event.Title
		}
	})

	// Phase 3: Multi-market concurrent stress.
	t.Run("concurrent_stress", func(t *testing.T) {
		var wg sync.WaitGroup
		errCh := make(chan error, 20)

		workers := []func(){
			func() { _, err := client.Markets(ctx, &types.GetMarketsParams{Limit: 5}); if err != nil { errCh <- err } },
			func() { _, err := client.ActiveMarkets(ctx); if err != nil { errCh <- err } },
			func() { _, err := client.Events(ctx, &types.GetEventsParams{Limit: 5}); if err != nil { errCh <- err } },
			func() { _, err := client.Search(ctx, &types.SearchParams{Q: "trump", LimitPerType: intPtr(3)}); if err != nil { errCh <- err } },
			func() { _, err := client.OrderBook(ctx, demoMarkets[0].tokens[0]); if err != nil { errCh <- err } },
			func() { _, err := client.Price(ctx, demoMarkets[0].tokens[0], "buy"); if err != nil { errCh <- err } },
			func() { _, err := client.Midpoint(ctx, demoMarkets[0].tokens[0]); if err != nil { errCh <- err } },
			func() { _, err := client.Spread(ctx, demoMarkets[0].tokens[0]); if err != nil { errCh <- err } },
			func() { _, err := client.TickSize(ctx, demoMarkets[0].tokens[0]); if err != nil { errCh <- err } },
			func() { _, err := client.LastTradePrice(ctx, demoMarkets[0].tokens[0]); if err != nil { errCh <- err } },
			func() { _, err := client.NegRisk(ctx, demoMarkets[0].tokens[0]); if err != nil { errCh <- err } },
			func() { _, err := client.FeeRateBps(ctx, demoMarkets[0].tokens[0]); if err != nil { errCh <- err } },
			func() { _, err := client.CLOBMarket(ctx, demoMarkets[0].conditionID); if err != nil { errCh <- err } },
			func() { _, err := client.CLOBMarketByToken(ctx, demoMarkets[0].tokens[0]); if err != nil { errCh <- err } },
			func() { _, err := client.HealthCheck(ctx); if err != nil { errCh <- err } },
		}

		for _, w := range workers {
			wg.Add(1)
			go func(fn func()) { defer wg.Done(); fn() }(w)
		}
		wg.Wait()
		close(errCh)

		var errs []error
		for err := range errCh {
			errs = append(errs, err)
		}
		report.concurrentErrors = len(errs)
		report.concurrentTotal = len(workers)
		if len(errs) > 0 {
			t.Logf("Concurrent: %d/%d failed", len(errs), len(workers))
			for _, err := range errs {
				t.Logf("  - %v", err)
			}
		} else {
			t.Logf("Concurrent: all %d succeeded", len(workers))
		}
	})

	// Phase 4: Error injection.
	t.Run("error_injection", func(t *testing.T) {
		cases := []struct {
			name string
			fn   func() error
		}{
			{"invalid_market", func() error { _, err := client.MarketByID(ctx, "bad"); return err }},
			{"invalid_token", func() error { _, err := client.OrderBook(ctx, "0xbad"); return err }},
			{"invalid_condition", func() error { _, err := client.CLOBMarket(ctx, "0xdead"); return err }},
			{"empty_search", func() error { _, err := client.Search(ctx, &types.SearchParams{Q: ""}); return err }},
			{"nonexistent_slug", func() error { _, err := client.MarketBySlug(ctx, "does-not-exist-123"); return err }},
		}

		for _, c := range cases {
			err := c.fn()
			if err == nil {
				t.Errorf("%s: expected error, got nil", c.name)
			} else {
				t.Logf("Error injection %s: %v", c.name, err)
			}
		}
	})

	// Phase 5: Consistency checks.
	t.Run("consistency", func(t *testing.T) {
		token := demoMarkets[0].tokens[0]

		// Midpoint consistency.
		mid1, _ := client.Midpoint(ctx, token)
		mid2, _ := client.Midpoint(ctx, token)
		if mid1 != mid2 {
			t.Logf("Midpoint drift: %s vs %s", mid1, mid2)
		}

		// Token resolution round-trip.
		clobByToken, _ := client.CLOBMarketByToken(ctx, token)
		if clobByToken != nil && clobByToken.ConditionID != demoMarkets[0].conditionID {
			t.Errorf("Token round-trip failed: got %s, want %s",
				clobByToken.ConditionID, demoMarkets[0].conditionID)
		}

		// Market active status.
		market, _ := client.MarketByID(ctx, demoMarkets[0].id)
		if market != nil && !market.Active {
			t.Errorf("Market %s is not active", demoMarkets[0].id)
		}
	})

	// Final summary.
	report.duration = time.Since(report.start)
	t.Logf("\n%s", report.String())
}

// demoReport collects structured results for the final summary.
type demoReport struct {
	start              time.Time
	health             string
	discoveredMarkets  int
	discoveredEvents   int
	discoveredTags     int
	discoveredSeries   int
	marketName         string
	marketPrices       string
	liquidity          float64
	volume             float64
	bids               int
	asks               int
	priceBuy           string
	priceSell          string
	midpoint           string
	spread             string
	lastTrade          string
	tickSize           string
	feeBps             int
	negRisk            bool
	acceptingOrders    bool
	conditionID        string
	priceHistoryPoints int
	comments           int
	searchResults      int
	simplifiedMarkets  int
	serverTime         string
	concurrentTotal    int
	concurrentErrors   int
	batchBooks         int
	batchPrices        int
	clobMarkets        int
	samplingMarkets    int
	leaderboard        int
	openInterest       float64
	topHolders         int
	liveVolume         float64
	eventTitle         string
	duration           time.Duration
}

func (r *demoReport) String() string {
	var b strings.Builder
	b.WriteString("╔══════════════════════════════════════════════════════════════╗\n")
	b.WriteString("║           POLYGOLEM LIVE E2E DEMO SUMMARY                    ║\n")
	b.WriteString("╠══════════════════════════════════════════════════════════════╣\n")
	b.WriteString(fmt.Sprintf("║ Health: %-52s ║\n", r.health))
	b.WriteString(fmt.Sprintf("║ Discovery: %d markets, %d events, %d tags, %d series       ║\n",
		r.discoveredMarkets, r.discoveredEvents, r.discoveredTags, r.discoveredSeries))
	b.WriteString("╠══════════════════════════════════════════════════════════════╣\n")
	b.WriteString(fmt.Sprintf("║ Market: %-53s║\n", truncate(r.marketName, 50)))
	b.WriteString(fmt.Sprintf("║ Prices: %-52s ║\n", r.marketPrices))
	b.WriteString(fmt.Sprintf("║ Liquidity: %.2f | Volume: %.2f                      ║\n", r.liquidity, r.volume))
	b.WriteString(fmt.Sprintf("║ Order Book: %d bids, %d asks                                ║\n", r.bids, r.asks))
	b.WriteString(fmt.Sprintf("║ Buy: %s | Sell: %s | Mid: %s | Spread: %s    ║\n",
		r.priceBuy, r.priceSell, r.midpoint, r.spread))
	b.WriteString(fmt.Sprintf("║ Last Trade: %-47s ║\n", r.lastTrade))
	b.WriteString(fmt.Sprintf("║ Tick: %s | Fee: %d bps | NegRisk: %v                        ║\n",
		r.tickSize, r.feeBps, r.negRisk))
	b.WriteString(fmt.Sprintf("║ Accepting Orders: %-39v ║\n", r.acceptingOrders))
	b.WriteString(fmt.Sprintf("║ Condition ID: %-45s ║\n", truncate(r.conditionID, 43)))
	b.WriteString("╠══════════════════════════════════════════════════════════════╣\n")
	b.WriteString(fmt.Sprintf("║ Price History: %d points                                    ║\n", r.priceHistoryPoints))
	b.WriteString(fmt.Sprintf("║ Comments: %d | Search Results: %d                           ║\n", r.comments, r.searchResults))
	b.WriteString(fmt.Sprintf("║ Simplified Markets: %d                                      ║\n", r.simplifiedMarkets))
	b.WriteString(fmt.Sprintf("║ Server Time: %-46s ║\n", r.serverTime))
	b.WriteString("╠══════════════════════════════════════════════════════════════╣\n")
	b.WriteString(fmt.Sprintf("║ Batch: %d books, %d prices                                  ║\n", r.batchBooks, r.batchPrices))
	b.WriteString(fmt.Sprintf("║ CLOB Markets: %d | Sampling: %d                             ║\n", r.clobMarkets, r.samplingMarkets))
	b.WriteString(fmt.Sprintf("║ Leaderboard: %d | Open Interest: %.2f                       ║\n", r.leaderboard, r.openInterest))
	b.WriteString(fmt.Sprintf("║ Top Holders: %d | Live Volume: %.2f                         ║\n", r.topHolders, r.liveVolume))
	b.WriteString(fmt.Sprintf("║ Event: %-51s║\n", truncate(r.eventTitle, 48)))
	b.WriteString("╠══════════════════════════════════════════════════════════════╣\n")
	b.WriteString(fmt.Sprintf("║ Concurrent: %d/%d succeeded                                  ║\n",
		r.concurrentTotal-r.concurrentErrors, r.concurrentTotal))
	b.WriteString(fmt.Sprintf("║ Duration: %-49s ║\n", r.duration))
	b.WriteString("╚══════════════════════════════════════════════════════════════╝")
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}


