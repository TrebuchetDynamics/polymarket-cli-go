package tests

import (
	"context"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/polygolem/pkg/types"
	"github.com/TrebuchetDynamics/polygolem/pkg/universal"
)

func TestRealMarketExtractionE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := universal.NewClient(universal.Config{})

	const (
		marketID    = "540817"
		tokenYes    = "98022490269692409998126496127597032490334070080325855126491859374983463996227"
		tokenNo     = "53831553061883006530739877284105938919721408776239639687877978808906551086026"
		conditionID = "0x1fad72fae204143ff1c3035e99e7c0f65ea8d5cd9bd1070987bd1a3316f772be"
		slug        = "new-rhianna-album-before-gta-vi-926"
	)

	t.Run("gamma_market_by_id", func(t *testing.T) {
		m, err := client.MarketByID(ctx, marketID)
		if err != nil {
			t.Fatalf("MarketByID: %v", err)
		}
		if m.ID != marketID {
			t.Fatalf("market id mismatch: got %s, want %s", m.ID, marketID)
		}
		if !m.Active {
			t.Fatal("market is not active")
		}
		if m.Closed {
			t.Fatal("market is unexpectedly closed")
		}
		if m.EnableOrderBook != true {
			t.Fatal("market does not have order book enabled")
		}
		if len(m.Outcomes) != 2 {
			t.Fatalf("expected 2 outcomes, got %d", len(m.Outcomes))
		}
		t.Logf("Market: %s | Outcomes: %v | Prices: %v | Liquidity: %f | Volume: %f",
			m.Question, m.Outcomes, m.OutcomePrices, m.LiquidityNum, m.VolumeNum)
	})

	t.Run("gamma_market_by_slug", func(t *testing.T) {
		m, err := client.MarketBySlug(ctx, slug)
		if err != nil {
			t.Logf("MarketBySlug returned error (slug endpoint may be restricted): %v", err)
			return
		}
		if m.ID != marketID {
			t.Fatalf("slug resolved to wrong market: got %s, want %s", m.ID, marketID)
		}
	})

	t.Run("gamma_market_by_token", func(t *testing.T) {
		m, err := client.MarketByToken(ctx, tokenYes)
		if err != nil {
			t.Logf("MarketByToken returned error (token endpoint may be restricted): %v", err)
			return
		}
		if m.Market.ID != marketID {
			t.Fatalf("token resolved to wrong market: got %s, want %s", m.Market.ID, marketID)
		}
	})

	t.Run("clob_order_book", func(t *testing.T) {
		book, err := client.OrderBook(ctx, tokenYes)
		if err != nil {
			t.Fatalf("OrderBook: %v", err)
		}
		if book.Market != conditionID {
			t.Fatalf("book market mismatch: got %s, want %s", book.Market, conditionID)
		}
		if book.AssetID != tokenYes {
			t.Fatalf("book asset_id mismatch: got %s, want %s", book.AssetID, tokenYes)
		}
		t.Logf("Bids: %d | Asks: %d | Hash: %s", len(book.Bids), len(book.Asks), book.Hash)
	})

	t.Run("clob_price", func(t *testing.T) {
		price, err := client.Price(ctx, tokenYes, "buy")
		if err != nil {
			t.Fatalf("Price: %v", err)
		}
		if price == "" {
			t.Fatal("price is empty")
		}
		t.Logf("Best buy price: %s", price)
	})

	t.Run("clob_spread", func(t *testing.T) {
		spread, err := client.Spread(ctx, tokenYes)
		if err != nil {
			t.Fatalf("Spread: %v", err)
		}
		if spread == "" {
			t.Fatal("spread is empty")
		}
		t.Logf("Spread: %s", spread)
	})

	t.Run("clob_midpoint", func(t *testing.T) {
		mid, err := client.Midpoint(ctx, tokenYes)
		if err != nil {
			t.Fatalf("Midpoint: %v", err)
		}
		if mid == "" {
			t.Fatal("midpoint is empty")
		}
		t.Logf("Midpoint: %s", mid)
	})

	t.Run("clob_tick_size", func(t *testing.T) {
		tick, err := client.TickSize(ctx, tokenYes)
		if err != nil {
			t.Fatalf("TickSize: %v", err)
		}
		if tick.MinimumTickSize == "" {
			t.Fatal("minimum tick size is empty")
		}
		t.Logf("MinTick: %s | MinSize: %s", tick.MinimumTickSize, tick.MinimumOrderSize)
	})

	t.Run("gamma_search", func(t *testing.T) {
		limit := 10
		res, err := client.Search(ctx, &types.SearchParams{Q: "Rihanna", LimitPerType: &limit})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		found := false
		for _, ev := range res.Events {
			if ev.ID == "23784" {
				found = true
				break
			}
		}
		if !found {
			t.Log("Note: parent event not found in search results (search index may lag)")
		}
	})

	t.Run("gamma_markets_list", func(t *testing.T) {
		markets, err := client.Markets(ctx, &types.GetMarketsParams{Limit: 50})
		if err != nil {
			t.Fatalf("Markets: %v", err)
		}
		found := false
		for _, m := range markets {
			if m.ID == marketID {
				found = true
				break
			}
		}
		if !found {
			t.Log("Note: market not in first 50 active markets (may be beyond limit)")
		}
	})

	t.Run("coherence_midpoint", func(t *testing.T) {
		mid1, err := client.Midpoint(ctx, tokenYes)
		if err != nil {
			t.Fatalf("Midpoint: %v", err)
		}
		mid2, err := client.Midpoint(ctx, tokenYes)
		if err != nil {
			t.Fatalf("Midpoint (2nd call): %v", err)
		}
		if mid1 != mid2 {
			t.Logf("midpoint drift between calls: %s vs %s", mid1, mid2)
		}
	})

	t.Run("clob_market_by_token", func(t *testing.T) {
		mt, err := client.CLOBMarketByToken(ctx, tokenYes)
		if err != nil {
			t.Fatalf("CLOBMarketByToken: %v", err)
		}
		if mt.ConditionID != conditionID {
			t.Fatalf("condition_id mismatch: got %s, want %s", mt.ConditionID, conditionID)
		}
		t.Logf("ConditionID: %s | PrimaryToken: %s | SecondaryToken: %s",
			mt.ConditionID, mt.PrimaryTokenID, mt.SecondaryTokenID)
	})

	t.Run("gamma_comments", func(t *testing.T) {
		entityType := "Event"
		eventID := 23784
		comments, err := client.Comments(ctx, &types.CommentQuery{
			EntityID:   &eventID,
			EntityType: &entityType,
			Limit:      5,
		})
		if err != nil {
			t.Logf("Comments returned error (endpoint may be restricted): %v", err)
			return
		}
		t.Logf("Comments: %d", len(comments))
		if len(comments) > 0 {
			c := comments[0]
			t.Logf("Latest comment: id=%s body=%q created=%s",
				c.ID, c.Body, c.CreatedAt)
		}
	})
}
