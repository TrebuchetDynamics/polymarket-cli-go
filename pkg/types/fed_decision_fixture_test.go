package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const fedDecisionFixtureDir = "../../fixtures/polymarket/events/fed-decision-in-june-825"

func TestFedDecisionGammaFixturePreservesMarketTruth(t *testing.T) {
	var event Event
	readJSONFixture(t, "gamma-event.json", &event)

	if event.ID != "101772" || event.Slug != "fed-decision-in-june-825" || event.Title != "Fed Decision in June?" {
		t.Fatalf("unexpected event identity: id=%q slug=%q title=%q", event.ID, event.Slug, event.Title)
	}
	if !event.Active || event.Closed || event.Archived {
		t.Fatalf("unexpected event state: active=%v closed=%v archived=%v", event.Active, event.Closed, event.Archived)
	}
	if got := event.EndDate.Time().Format(time.RFC3339); got != "2026-06-17T00:00:00Z" {
		t.Fatalf("endDate=%s", got)
	}
	if event.Volume <= 0 || event.Liquidity <= 0 {
		t.Fatalf("missing event metrics: volume=%f liquidity=%f", event.Volume, event.Liquidity)
	}
	if !hasTag(event.Tags, "economic-policy") || !hasTag(event.Tags, "politics") {
		t.Fatalf("missing expected event tags: %+v", event.Tags)
	}
	if len(event.Markets) != 5 {
		t.Fatalf("markets=%d want 5", len(event.Markets))
	}

	expectedSlugs := map[string]bool{
		"will-the-fed-decrease-interest-rates-by-25-bps-after-the-june-2026-meeting": false,
		"will-the-fed-increase-interest-rates-by-25-bps-after-the-june-2026-meeting": false,
		"will-the-fed-decrease-interest-rates-by-50-bps-after-the-june-2026-meeting": false,
		"will-there-be-no-change-in-fed-interest-rates-after-the-june-2026-meeting":  false,
		"will-the-fed-increase-interest-rates-by-50-bps-after-the-june-2026-meeting": false,
	}
	for _, market := range event.Markets {
		expectedSlugs[market.Slug] = true
		assertFedDecisionGammaMarket(t, market)
	}
	for slug, seen := range expectedSlugs {
		if !seen {
			t.Fatalf("missing expected market slug %q", slug)
		}
	}
}

func TestFedDecisionSearchFixtureFindsEventBeforeDetail(t *testing.T) {
	var search SearchResponse
	readJSONFixture(t, "gamma-search.json", &search)

	if search.Pagination.TotalResults < 1 {
		t.Fatalf("search pagination did not report results: %+v", search.Pagination)
	}
	event, ok := findEventBySlug(search.Events, "fed-decision-in-june-825")
	if !ok {
		t.Fatalf("search did not find fed-decision-in-june-825; events=%+v", search.Events)
	}
	if event.ID != "101772" || event.Title != "Fed Decision in June?" {
		t.Fatalf("unexpected search event identity: id=%q title=%q", event.ID, event.Title)
	}
	if len(event.Markets) != 5 {
		t.Fatalf("search event markets=%d want 5", len(event.Markets))
	}
	for _, market := range event.Markets {
		assertFedDecisionGammaMarket(t, market)
	}
}

func TestFedDecisionCLOBFixturesPreserveCurrentProviderShape(t *testing.T) {
	var markets []CLOBMarket
	readJSONFixture(t, "clob-markets.json", &markets)
	if len(markets) != 5 {
		t.Fatalf("clob markets=%d want 5", len(markets))
	}
	for _, market := range markets {
		if !strings.HasPrefix(market.ConditionID, "0x") {
			t.Fatalf("condition id was not parsed from abbreviated CLOB field: %+v", market)
		}
		if len(market.Tokens) != 2 {
			t.Fatalf("tokens=%d want 2 for %s", len(market.Tokens), market.ConditionID)
		}
		if market.Tokens[0].TokenID == "" || market.Tokens[0].Outcome == "" {
			t.Fatalf("token was not parsed from abbreviated CLOB field: %+v", market.Tokens[0])
		}
		if !market.AcceptingOrders || !market.EnableOrderBook || !market.NegRisk {
			t.Fatalf("current CLOB flags were not parsed: %+v", market)
		}
		if market.OrderMinSize <= 0 || market.OrderPriceMinTickSize <= 0 {
			t.Fatalf("current CLOB order constraints were not parsed: %+v", market)
		}
		if market.MakerBaseFee == 0 || market.TakerBaseFee == 0 {
			t.Fatalf("current CLOB fees were not parsed: %+v", market)
		}
		if market.FeeDetails.Rate <= 0 || market.FeeDetails.Exponent <= 0 || !market.FeeDetails.TakerOnly {
			t.Fatalf("current CLOB fee details were not parsed: %+v", market.FeeDetails)
		}
		if market.RewardsMinSize <= 0 || market.RewardsMaxSpread <= 0 || market.MinimumOrderAge <= 0 {
			t.Fatalf("current CLOB rewards/order-age fields were not parsed: %+v", market)
		}
	}

	var books []CLOBOrderBook
	readJSONFixture(t, "clob-books.json", &books)
	if len(books) != 10 {
		t.Fatalf("books=%d want 10", len(books))
	}
	for _, book := range books {
		if book.Market == "" || book.AssetID == "" || book.Timestamp == "" || book.Hash == "" {
			t.Fatalf("book identity missing: %+v", book)
		}
		if len(book.Bids) == 0 || len(book.Asks) == 0 {
			t.Fatalf("book depth missing for %s", book.AssetID)
		}
		if book.MinOrderSize == "" || book.TickSize == "" || book.LastTradePrice == "" {
			t.Fatalf("book metadata missing: %+v", book)
		}
		if !book.NegRisk {
			t.Fatalf("book neg-risk flag missing for %s", book.AssetID)
		}
	}
}

func assertFedDecisionGammaMarket(t *testing.T, market Market) {
	t.Helper()
	if market.ID == "" || market.Slug == "" || market.Question == "" || market.ConditionID == "" {
		t.Fatalf("market identity missing: %+v", market)
	}
	if !strings.HasPrefix(market.ConditionID, "0x") {
		t.Fatalf("condition id=%q", market.ConditionID)
	}
	if !market.Active || market.Closed || market.Archived || !market.AcceptingOrders || !market.EnableOrderBook {
		t.Fatalf("unexpected market state: %+v", market)
	}
	if got := market.EndDate.Time().Format(time.RFC3339); got != "2026-06-17T00:00:00Z" {
		t.Fatalf("%s endDate=%s", market.Slug, got)
	}
	if market.VolumeNum <= 0 || market.LiquidityNum <= 0 {
		t.Fatalf("market metrics missing: %+v", market)
	}
	if market.BestBid < 0 || market.BestBid > 1 || market.BestAsk < 0 || market.BestAsk > 1 {
		t.Fatalf("market bid/ask out of range: bid=%f ask=%f", market.BestBid, market.BestAsk)
	}
	if len(market.Outcomes) != 2 || market.Outcomes[0] != "Yes" || market.Outcomes[1] != "No" {
		t.Fatalf("outcomes=%+v", market.Outcomes)
	}
	if len(market.OutcomePrices) != 2 {
		t.Fatalf("outcome prices=%+v", market.OutcomePrices)
	}
	for _, rawPrice := range market.OutcomePrices {
		price, err := strconv.ParseFloat(rawPrice, 64)
		if err != nil || price < 0 || price > 1 {
			t.Fatalf("invalid outcome price %q", rawPrice)
		}
	}
	var tokenIDs []string
	if err := json.Unmarshal([]byte(market.ClobTokenIDs), &tokenIDs); err != nil {
		t.Fatalf("clobTokenIds=%q: %v", market.ClobTokenIDs, err)
	}
	if len(tokenIDs) != 2 || tokenIDs[0] == "" || tokenIDs[1] == "" {
		t.Fatalf("token ids=%+v", tokenIDs)
	}
}

func readJSONFixture(t *testing.T, name string, target any) {
	t.Helper()
	path := filepath.Join(fedDecisionFixtureDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("decode fixture %s: %v", path, err)
	}
}

func hasTag(tags []Tag, slug string) bool {
	for _, tag := range tags {
		if tag.Slug == slug {
			return true
		}
	}
	return false
}

func findEventBySlug(events []Event, slug string) (Event, bool) {
	for _, event := range events {
		if event.Slug == slug {
			return event, true
		}
	}
	return Event{}, false
}
