package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/polygolem/pkg/types"
	"github.com/TrebuchetDynamics/polygolem/pkg/universal"
)

var marketsToTest = []struct {
	id          string
	slug        string
	name        string
	category    string
	tokenYes    string
	tokenNo     string
	conditionID string
	eventID     int
}{
	{
		id:          "540817",
		slug:        "new-rhianna-album-before-gta-vi-926",
		name:        "Rihanna Album before GTA VI",
		category:    "entertainment",
		tokenYes:    "98022490269692409998126496127597032490334070080325855126491859374983463996227",
		tokenNo:     "53831553061883006530739877284105938919721408776239639687877978808906551086026",
		conditionID: "0x1fad72fae204143ff1c3035e99e7c0f65ea8d5cd9bd1070987bd1a3316f772be",
		eventID:     23784,
	},
}

func TestMultiMarketFullExtractionE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live E2E stress test in short mode")
	}

	client := universal.NewClient(universal.Config{})

	t.Run("phase1_discover", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		markets, err := client.Markets(ctx, &types.GetMarketsParams{
			Limit:  20,
			Active: boolPtr(true),
		})
		if err != nil {
			t.Fatalf("Markets: %v", err)
		}
		t.Logf("Discovered %d active markets", len(markets))

		events, err := client.Events(ctx, &types.GetEventsParams{
			Limit:    10,
			Featured: boolPtr(true),
		})
		if err != nil {
			t.Logf("Events error: %v", err)
		} else {
			t.Logf("Discovered %d featured events", len(events))
		}

		tags, err := client.Tags(ctx, &types.GetTagsParams{Limit: 10})
		if err != nil {
			t.Logf("Tags error: %v", err)
		} else {
			t.Logf("Discovered %d tags", len(tags))
			if len(tags) > 0 {
				t.Logf("First tag: %s (id=%s)", tags[0].Label, tags[0].ID)
			}
		}

		series, err := client.Series(ctx, &types.GetSeriesParams{Limit: 10})
		if err != nil {
			t.Logf("Series error: %v", err)
		} else {
			t.Logf("Discovered %d series", len(series))
		}

		sports, err := client.SportsMetadata(ctx)
		if err != nil {
			t.Logf("SportsMetadata error: %v", err)
		} else {
			t.Logf("Discovered %d sports", len(sports))
		}

		search, err := client.Search(ctx, &types.SearchParams{
			Q:              "crypto",
			LimitPerType:   intPtr(5),
			SearchProfiles: boolPtr(false),
		})
		if err != nil {
			t.Logf("Search error: %v", err)
		} else {
			t.Logf("Search returned %d events", len(search.Events))
		}

		enriched, err := client.EnrichedMarkets(ctx, 5)
		if err != nil {
			t.Logf("EnrichedMarkets error: %v", err)
		} else {
			t.Logf("Enriched %d markets", len(enriched))
		}

		if len(markets) > 0 && len(markets[0].ClobTokenIDs) > 0 {
			t.Logf("First market tokens: %v", markets[0].ClobTokenIDs)
		}
	})

	t.Run("phase2_deep_extraction", func(t *testing.T) {
		m := marketsToTest[0]
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		gammaMarket, err := client.MarketByID(ctx, m.id)
		if err != nil {
			t.Fatalf("MarketByID: %v", err)
		}
		t.Logf("Gamma: %s | liquidity=%.2f | volume=%.2f",
			gammaMarket.Question, gammaMarket.LiquidityNum, gammaMarket.VolumeNum)

		book, err := client.OrderBook(ctx, m.tokenYes)
		if err != nil {
			t.Fatalf("OrderBook: %v", err)
		}
		t.Logf("Book: bids=%d asks=%d hash=%s", len(book.Bids), len(book.Asks), book.Hash)

		priceBuy, _ := client.Price(ctx, m.tokenYes, "buy")
		priceSell, _ := client.Price(ctx, m.tokenYes, "sell")
		mid, _ := client.Midpoint(ctx, m.tokenYes)
		spread, _ := client.Spread(ctx, m.tokenYes)
		lastTrade, _ := client.LastTradePrice(ctx, m.tokenYes)
		t.Logf("Prices: buy=%s sell=%s mid=%s spread=%s last=%s",
			priceBuy, priceSell, mid, spread, lastTrade)

		tick, _ := client.TickSize(ctx, m.tokenYes)
		fee, _ := client.FeeRateBps(ctx, m.tokenYes)
		negRisk, _ := client.NegRisk(ctx, m.tokenYes)
		t.Logf("Tick: min=%s | Fee: %d bps | NegRisk: %v",
			tick.MinimumTickSize, fee, negRisk.NegRisk)

		clobMkt, _ := client.CLOBMarket(ctx, m.conditionID)
		if clobMkt != nil {
			t.Logf("CLOBMarket: spread=%.4f accepting=%v", clobMkt.Spread, clobMkt.AcceptingOrders)
		}

		clobByToken, _ := client.CLOBMarketByToken(ctx, m.tokenYes)
		if clobByToken != nil {
			t.Logf("CLOBMarketByToken: condition=%s", clobByToken.ConditionID)
		}

		endTS := time.Now().Unix()
		startTS := endTS - 3600
		hist, err := client.PricesHistory(ctx, &types.CLOBPriceHistoryParams{
			Market:   m.conditionID,
			StartTS:  startTS,
			EndTS:    endTS,
			Fidelity: 60,
		})
		if err != nil {
			t.Logf("PricesHistory error: %v", err)
		} else {
			t.Logf("Price history: %d points", len(hist.History))
		}

		rewards, _ := client.RewardsConfig(ctx)
		t.Logf("Rewards configs: %d", len(rewards))

		entityType := "Event"
		comments, _ := client.Comments(ctx, &types.CommentQuery{
			EntityID:   &m.eventID,
			EntityType: &entityType,
			Limit:      3,
		})
		t.Logf("Comments: %d", len(comments))

		serverTime, _ := client.CLOBServerTime(ctx)
		if serverTime != nil {
			t.Logf("CLOB server time: %s", serverTime.ISO)
		}

		simplified, _ := client.SimplifiedMarkets(ctx, "")
		if simplified != nil {
			t.Logf("SimplifiedMarkets: %d markets, nextCursor=%s",
				len(simplified.Data), simplified.NextCursor)
		}
	})

	t.Run("phase3_concurrent_stress", func(t *testing.T) {
		m := marketsToTest[0]
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var wg sync.WaitGroup
		errors := make(chan error, 20)

		workers := []func(){
			func() { _, err := client.MarketByID(ctx, m.id); if err != nil { errors <- err } },
			func() { _, err := client.OrderBook(ctx, m.tokenYes); if err != nil { errors <- err } },
			func() { _, err := client.Price(ctx, m.tokenYes, "buy"); if err != nil { errors <- err } },
			func() { _, err := client.Midpoint(ctx, m.tokenYes); if err != nil { errors <- err } },
			func() { _, err := client.Spread(ctx, m.tokenYes); if err != nil { errors <- err } },
			func() { _, err := client.TickSize(ctx, m.tokenYes); if err != nil { errors <- err } },
			func() { _, err := client.LastTradePrice(ctx, m.tokenYes); if err != nil { errors <- err } },
			func() { _, err := client.NegRisk(ctx, m.tokenYes); if err != nil { errors <- err } },
			func() { _, err := client.CLOBMarketByToken(ctx, m.tokenYes); if err != nil { errors <- err } },
			func() { _, err := client.CLOBMarket(ctx, m.conditionID); if err != nil { errors <- err } },
			func() { _, err := client.Markets(ctx, &types.GetMarketsParams{Limit: 5}); if err != nil { errors <- err } },
			func() { _, err := client.Search(ctx, &types.SearchParams{Q: "trump", LimitPerType: intPtr(3)}); if err != nil { errors <- err } },
			func() { _, err := client.ActiveMarkets(ctx); if err != nil { errors <- err } },
			func() { _, err := client.HealthCheck(ctx); if err != nil { errors <- err } },
			func() { _, err := client.Events(ctx, &types.GetEventsParams{Limit: 5}); if err != nil { errors <- err } },
		}

		for _, w := range workers {
			wg.Add(1)
			go func(fn func()) {
				defer wg.Done()
				fn()
			}(w)
		}

		wg.Wait()
		close(errors)

		errCount := 0
		for err := range errors {
			t.Logf("Concurrent error: %v", err)
			errCount++
		}
		if errCount > 0 {
			t.Logf("Concurrent stress: %d/%d requests failed", errCount, len(workers))
		} else {
			t.Logf("Concurrent stress: all %d requests succeeded", len(workers))
		}
	})

	t.Run("phase4_error_injection", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		_, err := client.MarketByID(ctx, "not-an-id")
		if err == nil {
			t.Fatal("expected error for invalid market ID")
		}
		t.Logf("Invalid market ID: %v", err)

		_, err = client.OrderBook(ctx, "0xinvalid")
		if err == nil {
			t.Fatal("expected error for invalid token ID")
		}
		t.Logf("Invalid token ID: %v", err)

		_, err = client.Search(ctx, &types.SearchParams{Q: "", LimitPerType: intPtr(1)})
		if err != nil {
			t.Logf("Empty search error (may be ok): %v", err)
		}

		_, err = client.CLOBMarket(ctx, "0xdeadbeef")
		if err == nil {
			t.Fatal("expected error for invalid condition ID")
		}
		t.Logf("Invalid condition ID: %v", err)

		_, err = client.MarketBySlug(ctx, "this-market-definitely-does-not-exist-12345")
		if err == nil {
			t.Fatal("expected error for nonexistent slug")
		}
		t.Logf("Nonexistent slug: %v", err)
	})

	t.Run("phase5_keyset_pagination", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		markets, cursor, err := client.MarketsKeyset(ctx, &types.KeysetParams{
			Limit: 5,
		})
		if err != nil {
			t.Logf("MarketsKeyset error (endpoint may be deprecated): %v", err)
			return
		}
		t.Logf("Keyset page 1: %d markets, nextCursor=%s", len(markets), cursor)

		if cursor != "" {
			markets2, cursor2, err := client.MarketsKeyset(ctx, &types.KeysetParams{
				Limit:    5,
				KeysetID: cursor,
			})
			if err != nil {
				t.Logf("Keyset page 2 error: %v", err)
			} else {
				t.Logf("Keyset page 2: %d markets, nextCursor=%s", len(markets2), cursor2)
			}
		}
	})
}

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }
