package tests

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/polygolem/pkg/stream"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
	"github.com/TrebuchetDynamics/polygolem/pkg/universal"
)

func TestPolygolemLiveReadOnlyE2E(t *testing.T) {
	if os.Getenv("POLYGOLEM_LIVE_READONLY_E2E") != "1" && os.Getenv("POLYGOLEM_LIVE_E2E") != "1" {
		t.Skip("set POLYGOLEM_LIVE_READONLY_E2E=1 to run live read-only Polymarket E2E")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := universal.NewClient(universal.DefaultConfig())

	health, err := client.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	if !health.GammaOK || !health.CLOBOK || !health.DataOK {
		t.Fatalf("health = %+v", health)
	}

	active := true
	closed := false
	desc := false
	markets, err := client.Markets(ctx, &types.GetMarketsParams{
		Limit:     20,
		Active:    &active,
		Closed:    &closed,
		Order:     "volume24hr",
		Ascending: &desc,
	})
	if err != nil {
		t.Fatalf("Markets: %v", err)
	}
	if len(markets) == 0 {
		t.Fatal("Markets returned no active markets")
	}
	gammaMarket, tokenIDs := chooseGammaMarketWithTokens(markets)
	if gammaMarket == nil {
		t.Fatalf("no active Gamma market had CLOB token IDs: got %d markets", len(markets))
	}
	if gammaMarket.Slug == "" || gammaMarket.ConditionID == "" {
		t.Fatalf("Gamma market missing slug or condition ID: %+v", *gammaMarket)
	}

	events, err := client.Events(ctx, &types.GetEventsParams{Limit: 10, Closed: &closed})
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("Events returned no active events")
	}

	tags, err := client.Tags(ctx, &types.GetTagsParams{Limit: 10})
	if err != nil {
		t.Fatalf("Tags: %v", err)
	}
	if len(tags) == 0 || tags[0].Slug == "" {
		t.Fatalf("Tags did not return usable tag taxonomy: %+v", tags)
	}

	series, err := client.Series(ctx, &types.GetSeriesParams{Limit: 10, Closed: &closed})
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	if len(series) == 0 {
		t.Fatal("Series returned no rows")
	}

	if countTaxonomy(markets, events, series) == 0 {
		t.Fatalf("live taxonomy had no categories/tags across markets/events/series")
	}

	searchQuery := firstSearchTerm(*gammaMarket)
	search, err := client.Search(ctx, &types.SearchParams{Q: searchQuery})
	if err != nil {
		t.Fatalf("Search(%q): %v", searchQuery, err)
	}
	if len(search.Events) == 0 && len(search.Tags) == 0 && len(search.Profiles) == 0 {
		t.Fatalf("Search(%q) returned no events, tags, or profiles", searchQuery)
	}

	if id, err := strconv.Atoi(events[0].ID); err == nil {
		entityType := "Event"
		if _, err := client.Comments(ctx, &types.CommentQuery{
			EntityID:   &id,
			EntityType: &entityType,
			Limit:      3,
		}); err != nil {
			t.Fatalf("Comments(event %s): %v", events[0].ID, err)
		}
	}

	clobMarkets, err := client.CLOBMarkets(ctx, "")
	if err != nil {
		t.Fatalf("CLOBMarkets: %v", err)
	}
	if clobMarkets == nil || len(clobMarkets.Data) == 0 {
		t.Fatalf("CLOBMarkets returned no rows: %+v", clobMarkets)
	}
	clobMarket := chooseCLOBMarket(clobMarkets.Data, gammaMarket.ConditionID)
	if clobMarket == nil {
		clobMarket = &clobMarkets.Data[0]
	}
	if len(clobMarket.Tokens) > 0 {
		tokenIDs = appendUnique(tokenIDs, clobMarket.Tokens[0].TokenID)
	}

	fetchedCLOBMarket, err := client.CLOBMarket(ctx, clobMarket.ConditionID)
	if err != nil {
		t.Fatalf("CLOBMarket(%q): %v", clobMarket.ConditionID, err)
	}
	if fetchedCLOBMarket.ConditionID == "" || len(fetchedCLOBMarket.Tokens) == 0 {
		t.Fatalf("CLOBMarket returned unusable market: %+v", fetchedCLOBMarket)
	}
	tokenIDs = appendUnique(tokenIDs, fetchedCLOBMarket.Tokens[0].TokenID)

	tokenID := firstNonEmptyString(tokenIDs...)
	if tokenID == "" {
		t.Fatalf("no CLOB token ID found from Gamma market %+v and CLOB market %+v", *gammaMarket, *clobMarket)
	}

	book, err := client.OrderBook(ctx, tokenID)
	if err != nil {
		t.Fatalf("OrderBook(%s): %v", tokenID, err)
	}
	if book.AssetID == "" || book.Market == "" {
		t.Fatalf("OrderBook returned missing identifiers: %+v", book)
	}
	if len(book.Bids) == 0 && len(book.Asks) == 0 {
		t.Fatalf("OrderBook(%s) has no bids or asks", tokenID)
	}

	if _, err := client.TickSize(ctx, tokenID); err != nil {
		t.Fatalf("TickSize(%s): %v", tokenID, err)
	}
	if _, err := client.NegRisk(ctx, tokenID); err != nil {
		t.Fatalf("NegRisk(%s): %v", tokenID, err)
	}
	if _, err := client.FeeRateBps(ctx, tokenID); err != nil {
		t.Fatalf("FeeRateBps(%s): %v", tokenID, err)
	}
	if _, err := client.Midpoint(ctx, tokenID); err != nil {
		t.Fatalf("Midpoint(%s): %v", tokenID, err)
	}
	if _, err := client.Spread(ctx, tokenID); err != nil {
		t.Fatalf("Spread(%s): %v", tokenID, err)
	}
	if _, err := client.Prices(ctx, []types.CLOBBookParams{{TokenID: tokenID, Side: "BUY"}, {TokenID: tokenID, Side: "SELL"}}); err != nil {
		t.Fatalf("Prices(%s): %v", tokenID, err)
	}
	if _, err := client.OrderBooks(ctx, []types.CLOBBookParams{{TokenID: tokenID}}); err != nil {
		t.Fatalf("OrderBooks(%s): %v", tokenID, err)
	}
	if _, err := client.PricesHistory(ctx, &types.CLOBPriceHistoryParams{Market: tokenID, Interval: "1h", Fidelity: 60}); err != nil {
		t.Fatalf("PricesHistory(%s): %v", tokenID, err)
	}

	if _, err := client.OpenInterest(ctx, gammaMarket.ConditionID); err != nil {
		t.Fatalf("OpenInterest(%s): %v", gammaMarket.ConditionID, err)
	}
	holders, err := client.TopHolders(ctx, gammaMarket.ConditionID, 3)
	if err != nil {
		t.Fatalf("TopHolders(%s): %v", gammaMarket.ConditionID, err)
	}
	eventID, err := strconv.Atoi(events[0].ID)
	if err != nil {
		t.Fatalf("event ID %q is not numeric for LiveVolume: %v", events[0].ID, err)
	}
	liveVolume, err := client.LiveVolume(ctx, eventID)
	if err != nil {
		t.Fatalf("LiveVolume(%d): %v", eventID, err)
	}
	if liveVolume == nil || liveVolume.Total == 0 || len(liveVolume.Markets) == 0 {
		t.Fatalf("LiveVolume returned no market volume: %+v", liveVolume)
	}
	leaderboard, err := client.TraderLeaderboard(ctx, 5)
	if err != nil {
		t.Fatalf("TraderLeaderboard: %v", err)
	}
	if len(leaderboard) == 0 {
		t.Fatal("TraderLeaderboard returned no rows")
	}

	userAddress := firstDataUser(holders, leaderboard)
	if userAddress != "" {
		if _, err := client.CurrentPositionsWithLimit(ctx, userAddress, 3); err != nil {
			t.Fatalf("CurrentPositionsWithLimit(%s): %v", userAddress, err)
		}
		if _, err := client.ClosedPositionsWithLimit(ctx, userAddress, 3); err != nil {
			t.Fatalf("ClosedPositionsWithLimit(%s): %v", userAddress, err)
		}
		if _, err := client.Trades(ctx, userAddress, 3); err != nil {
			t.Fatalf("Trades(%s): %v", userAddress, err)
		}
		if _, err := client.Activity(ctx, userAddress, 3); err != nil {
			t.Fatalf("Activity(%s): %v", userAddress, err)
		}
		if _, err := client.TotalValue(ctx, userAddress); err != nil {
			t.Fatalf("TotalValue(%s): %v", userAddress, err)
		}
		if _, err := client.MarketsTraded(ctx, userAddress); err != nil {
			t.Fatalf("MarketsTraded(%s): %v", userAddress, err)
		}
	}

	assertLiveMarketStream(t, ctx, tokenID)
}

func chooseGammaMarketWithTokens(markets []types.Market) (*types.Market, []string) {
	for i := range markets {
		ids := parseCLOBTokenIDs(markets[i].ClobTokenIDs)
		if markets[i].ConditionID != "" && len(ids) > 0 {
			return &markets[i], ids
		}
	}
	return nil, nil
}

func chooseCLOBMarket(markets []types.CLOBMarket, conditionID string) *types.CLOBMarket {
	for i := range markets {
		if markets[i].ConditionID == conditionID && len(markets[i].Tokens) > 0 {
			return &markets[i]
		}
	}
	for i := range markets {
		if len(markets[i].Tokens) > 0 {
			return &markets[i]
		}
	}
	return nil
}

func parseCLOBTokenIDs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(raw), &ids); err == nil {
		return cleanStrings(ids)
	}
	return cleanStrings(strings.Split(raw, ","))
}

func countTaxonomy(markets []types.Market, events []types.Event, series []types.Series) int {
	var count int
	for _, market := range markets {
		if market.Category != "" {
			count++
		}
		count += len(market.Categories) + len(market.Tags)
	}
	for _, event := range events {
		if event.Category != "" || event.Subcategory != "" {
			count++
		}
		count += len(event.Categories) + len(event.Tags)
	}
	for _, row := range series {
		count += len(row.Categories) + len(row.Tags)
	}
	return count
}

func firstSearchTerm(market types.Market) string {
	for _, source := range []string{market.Category, market.Slug, market.Question} {
		for _, field := range strings.FieldsFunc(source, func(r rune) bool {
			return r == '-' || r == '_' || r == '?' || r == ' '
		}) {
			field = strings.TrimSpace(field)
			if len(field) >= 4 && !strings.EqualFold(field, "will") {
				return field
			}
		}
	}
	return "crypto"
}

func firstDataUser(holders []types.Holder, leaderboard []types.LeaderboardRow) string {
	for _, holder := range holders {
		if strings.TrimSpace(holder.Address) != "" {
			return holder.Address
		}
	}
	for _, row := range leaderboard {
		if strings.TrimSpace(row.User) != "" {
			return row.User
		}
	}
	return ""
}

func assertLiveMarketStream(t *testing.T, ctx context.Context, tokenID string) {
	t.Helper()

	streamCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	cfg := stream.DefaultConfig("")
	cfg.Reconnect = false
	cfg.PingInterval = 30 * time.Second
	client := stream.NewMarketClient(cfg)
	defer client.Close()

	events := make(chan string, 1)
	errs := make(chan error, 1)
	client.OnBook = func(msg stream.BookMessage) {
		if msg.AssetID != "" || msg.Market != "" {
			sendStreamEvent(events, "book")
		}
	}
	client.OnPriceChange = func(msg stream.PriceChangeMessage) {
		if msg.Market != "" || len(msg.PriceChanges) > 0 {
			sendStreamEvent(events, "price_change")
		}
	}
	client.OnLastTrade = func(msg stream.LastTradeMessage) {
		if msg.AssetID != "" || msg.Market != "" {
			sendStreamEvent(events, "last_trade_price")
		}
	}
	client.OnError = func(err error) {
		select {
		case errs <- err:
		default:
		}
	}

	if err := client.Connect(streamCtx); err != nil {
		t.Fatalf("stream connect: %v", err)
	}
	if err := client.SubscribeAssets(streamCtx, []string{tokenID}); err != nil {
		t.Fatalf("stream subscribe %s: %v", tokenID, err)
	}
	select {
	case event := <-events:
		t.Logf("live stream received %s for token %s", event, tokenID)
	case err := <-errs:
		t.Fatalf("live stream error for token %s: %v", tokenID, err)
	case <-streamCtx.Done():
		t.Fatalf("live stream timed out waiting for token %s: %v", tokenID, streamCtx.Err())
	}
}

func sendStreamEvent(ch chan<- string, event string) {
	select {
	case ch <- event:
	default:
	}
}

func cleanStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.Trim(value, ` "'[]`)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func appendUnique(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
