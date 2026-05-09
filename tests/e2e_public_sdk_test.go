package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/polygolem/pkg/clob"
	"github.com/TrebuchetDynamics/polygolem/pkg/contracts"
	"github.com/TrebuchetDynamics/polygolem/pkg/data"
	"github.com/TrebuchetDynamics/polygolem/pkg/relayer"
	"github.com/TrebuchetDynamics/polygolem/pkg/settlement"
	"github.com/TrebuchetDynamics/polygolem/pkg/stream"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
	"github.com/TrebuchetDynamics/polygolem/pkg/universal"
	"github.com/gorilla/websocket"
)

const e2ePrivateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
const e2eBuilderCode = "0x1111111111111111111111111111111111111111111111111111111111111111"
const e2eConfiguredCLOBKey = "configured-key"
const e2eConfiguredCLOBSecret = "c2VjcmV0"
const e2eConfiguredCLOBPassphrase = "configured-pass"

func TestPolygolemPublicSDKE2EAgainstLocalPolymarket(t *testing.T) {
	rec := newE2ERecorder()
	gammaServer := newE2EGammaServer(t, rec)
	defer gammaServer.Close()
	clobServer := newE2ECLOBServer(t, rec)
	defer clobServer.Close()
	dataServer := newE2EDataServer(t, rec)
	defer dataServer.Close()
	relayerServer := newE2ERelayerServer(t, rec)
	defer relayerServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	authClient := universal.NewClient(universal.Config{CLOBBaseURL: clobServer.URL})
	createdKey, err := authClient.CreateOrDeriveAPIKey(ctx, e2ePrivateKey)
	if err != nil {
		t.Fatalf("CreateOrDeriveAPIKey: %v", err)
	}
	if createdKey.Key != "created-key" || createdKey.Passphrase != "created-pass" {
		t.Fatalf("created key = %+v", createdKey)
	}

	client := universal.NewClient(universal.Config{
		GammaBaseURL: gammaServer.URL,
		CLOBBaseURL:  clobServer.URL,
		DataBaseURL:  dataServer.URL,
		BuilderCode:  e2eBuilderCode,
		CLOBCredentials: clob.APIKey{
			Key:        e2eConfiguredCLOBKey,
			Secret:     e2eConfiguredCLOBSecret,
			Passphrase: e2eConfiguredCLOBPassphrase,
		},
	})

	health, err := client.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	if !health.GammaOK || !health.CLOBOK || !health.DataOK {
		t.Fatalf("health = %+v", health)
	}

	markets, err := client.Markets(ctx, &types.GetMarketsParams{Limit: 1})
	if err != nil {
		t.Fatalf("Markets: %v", err)
	}
	if len(markets) != 1 || markets[0].ClobTokenIDs == "" {
		t.Fatalf("markets = %+v", markets)
	}
	if markets[0].Category != "Crypto" || len(markets[0].Categories) != 1 || markets[0].Categories[0].Slug != "crypto" {
		t.Fatalf("market categories = category:%q categories:%+v", markets[0].Category, markets[0].Categories)
	}
	activeMarkets, err := client.ActiveMarkets(ctx)
	if err != nil {
		t.Fatalf("ActiveMarkets: %v", err)
	}
	if len(activeMarkets) != 1 || !activeMarkets[0].Active {
		t.Fatalf("active markets = %+v", activeMarkets)
	}
	marketByID, err := client.MarketByID(ctx, "market-1")
	if err != nil {
		t.Fatalf("MarketByID: %v", err)
	}
	if marketByID.ID != "market-1" {
		t.Fatalf("market by id = %+v", marketByID)
	}
	marketBySlug, err := client.MarketBySlug(ctx, "will-btc-hit-100k")
	if err != nil {
		t.Fatalf("MarketBySlug: %v", err)
	}
	if marketBySlug.Slug != "will-btc-hit-100k" {
		t.Fatalf("market by slug = %+v", marketBySlug)
	}
	events, err := client.Events(ctx, &types.GetEventsParams{Limit: 1})
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	if len(events) != 1 || len(events[0].Markets) != 1 {
		t.Fatalf("events = %+v", events)
	}
	if len(events[0].Categories) != 1 || events[0].Categories[0].Slug != "crypto" || len(events[0].Markets[0].Categories) != 1 {
		t.Fatalf("event categories = event:%+v market:%+v", events[0].Categories, events[0].Markets[0].Categories)
	}
	eventByID, err := client.EventByID(ctx, "event-1")
	if err != nil {
		t.Fatalf("EventByID: %v", err)
	}
	if eventByID.ID != "event-1" {
		t.Fatalf("event by id = %+v", eventByID)
	}
	eventBySlug, err := client.EventBySlug(ctx, "btc-event")
	if err != nil {
		t.Fatalf("EventBySlug: %v", err)
	}
	if eventBySlug.Slug != "btc-event" {
		t.Fatalf("event by slug = %+v", eventBySlug)
	}
	search, err := client.Search(ctx, &types.SearchParams{Q: "btc"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(search.Events) != 1 || search.Events[0].Slug != "btc-event" {
		t.Fatalf("search = %+v", search)
	}
	series, err := client.Series(ctx, &types.GetSeriesParams{Limit: 1})
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	if len(series) != 1 || series[0].Slug != "crypto" {
		t.Fatalf("series = %+v", series)
	}
	if len(series[0].Categories) != 1 || series[0].Categories[0].Slug != "crypto" {
		t.Fatalf("series categories = %+v", series[0].Categories)
	}
	seriesByID, err := client.SeriesByID(ctx, "series-1")
	if err != nil {
		t.Fatalf("SeriesByID: %v", err)
	}
	if seriesByID.ID != "series-1" {
		t.Fatalf("series by id = %+v", seriesByID)
	}
	tags, err := client.Tags(ctx, &types.GetTagsParams{Limit: 1})
	if err != nil {
		t.Fatalf("Tags: %v", err)
	}
	if len(tags) != 1 || tags[0].Slug != "crypto" {
		t.Fatalf("tags = %+v", tags)
	}
	tagByID, err := client.TagByID(ctx, "1")
	if err != nil {
		t.Fatalf("TagByID: %v", err)
	}
	if tagByID.ID != "1" {
		t.Fatalf("tag by id = %+v", tagByID)
	}
	tagBySlug, err := client.TagBySlug(ctx, "crypto")
	if err != nil {
		t.Fatalf("TagBySlug: %v", err)
	}
	if tagBySlug.Slug != "crypto" {
		t.Fatalf("tag by slug = %+v", tagBySlug)
	}
	related, err := client.RelatedTagsByID(ctx, "1")
	if err != nil {
		t.Fatalf("RelatedTagsByID: %v", err)
	}
	if len(related) != 1 || related[0].RelatedTagID != 2 {
		t.Fatalf("related tags by id = %+v", related)
	}
	relatedBySlug, err := client.RelatedTagsBySlug(ctx, "crypto")
	if err != nil {
		t.Fatalf("RelatedTagsBySlug: %v", err)
	}
	if len(relatedBySlug) != 1 || relatedBySlug[0].RelatedTagID != 2 {
		t.Fatalf("related tags by slug = %+v", relatedBySlug)
	}
	teams, err := client.Teams(ctx, &types.GetTeamsParams{Limit: 1})
	if err != nil {
		t.Fatalf("Teams: %v", err)
	}
	if len(teams) != 1 || teams[0].Name != "Yankees" {
		t.Fatalf("teams = %+v", teams)
	}
	sports, err := client.SportsMetadata(ctx)
	if err != nil {
		t.Fatalf("SportsMetadata: %v", err)
	}
	if len(sports) != 1 || sports[0].Sport != "mlb" {
		t.Fatalf("sports metadata = %+v", sports)
	}
	entityType := "market"
	entityID := 1
	comments, err := client.Comments(ctx, &types.CommentQuery{EntityID: &entityID, EntityType: &entityType, Limit: 2})
	if err != nil {
		t.Fatalf("Comments: %v", err)
	}
	if len(comments) != 1 || comments[0].Body != "local sentiment" {
		t.Fatalf("comments = %+v", comments)
	}
	comment, err := client.CommentByID(ctx, "comment-1")
	if err != nil {
		t.Fatalf("CommentByID: %v", err)
	}
	if comment.ID != "comment-1" {
		t.Fatalf("comment by id = %+v", comment)
	}
	userComments, err := client.CommentsByUser(ctx, "0xuser", 2)
	if err != nil {
		t.Fatalf("CommentsByUser: %v", err)
	}
	if len(userComments) != 1 || userComments[0].User.Address != "0xuser" {
		t.Fatalf("comments by user = %+v", userComments)
	}
	profile, err := client.PublicProfile(ctx, "0xuser")
	if err != nil {
		t.Fatalf("PublicProfile: %v", err)
	}
	if profile.ProxyWallet != "0xuser" {
		t.Fatalf("profile = %+v", profile)
	}
	sportsMarketTypes, err := client.SportsMarketTypes(ctx)
	if err != nil {
		t.Fatalf("SportsMarketTypes: %v", err)
	}
	if len(sportsMarketTypes) != 1 || sportsMarketTypes[0].Slug != "moneyline" {
		t.Fatalf("sports market types = %+v", sportsMarketTypes)
	}
	byToken, err := client.MarketByToken(ctx, "12345")
	if err != nil {
		t.Fatalf("MarketByToken: %v", err)
	}
	if byToken.TokenID != "12345" || byToken.Market.ID != "market-1" {
		t.Fatalf("market by token = %+v", byToken)
	}
	keysetEvents, nextEventCursor, err := client.EventsKeyset(ctx, &types.KeysetParams{Limit: 1})
	if err != nil {
		t.Fatalf("EventsKeyset: %v", err)
	}
	if len(keysetEvents) != 1 || nextEventCursor != "event-cursor" {
		t.Fatalf("events keyset = %+v next=%q", keysetEvents, nextEventCursor)
	}
	keysetMarkets, nextMarketCursor, err := client.MarketsKeyset(ctx, &types.KeysetParams{Limit: 1})
	if err != nil {
		t.Fatalf("MarketsKeyset: %v", err)
	}
	if len(keysetMarkets) != 1 || nextMarketCursor != "market-cursor" {
		t.Fatalf("markets keyset = %+v next=%q", keysetMarkets, nextMarketCursor)
	}
	enriched, err := client.EnrichedMarkets(ctx, 1)
	if err != nil {
		t.Fatalf("EnrichedMarkets: %v", err)
	}
	if len(enriched) != 1 || enriched[0].FeeRateBps != 30 || enriched[0].OrderBook.AssetID != "12345" {
		t.Fatalf("enriched markets = %+v", enriched)
	}
	searchEnriched, err := client.SearchAndEnrich(ctx, "btc", 1)
	if err != nil {
		t.Fatalf("SearchAndEnrich: %v", err)
	}
	if len(searchEnriched) != 1 || searchEnriched[0].Market.ID != "market-1" {
		t.Fatalf("search enriched = %+v", searchEnriched)
	}

	serverTime, err := client.CLOBServerTime(ctx)
	if err != nil {
		t.Fatalf("CLOBServerTime: %v", err)
	}
	if serverTime.Timestamp != "1710000000" {
		t.Fatalf("server time = %+v", serverTime)
	}
	clobMarkets, err := client.CLOBMarkets(ctx, "")
	if err != nil {
		t.Fatalf("CLOBMarkets: %v", err)
	}
	if clobMarkets.Count != 1 || clobMarkets.Data[0].Tokens[0].TokenID != "12345" {
		t.Fatalf("clob markets = %+v", clobMarkets)
	}
	clobMarket, err := client.CLOBMarket(ctx, "condition-1")
	if err != nil {
		t.Fatalf("CLOBMarket: %v", err)
	}
	if clobMarket.ConditionID != "condition-1" || clobMarket.OrderMinSize != 5 {
		t.Fatalf("clob market = %+v", clobMarket)
	}
	if simplified, err := client.SimplifiedMarkets(ctx, "LTE="); err != nil || simplified.Count != 1 {
		t.Fatalf("SimplifiedMarkets = %+v err=%v", simplified, err)
	}
	if sampling, err := client.SamplingMarkets(ctx, "LTE="); err != nil || sampling.Count != 1 {
		t.Fatalf("SamplingMarkets = %+v err=%v", sampling, err)
	}
	if samplingSimple, err := client.SamplingSimplifiedMarkets(ctx, "LTE="); err != nil || samplingSimple.Count != 1 {
		t.Fatalf("SamplingSimplifiedMarkets = %+v err=%v", samplingSimple, err)
	}
	book, err := client.OrderBook(ctx, "12345")
	if err != nil {
		t.Fatalf("OrderBook: %v", err)
	}
	if book.AssetID != "12345" || book.Bids[0].Price != "0.44" {
		t.Fatalf("order book = %+v", book)
	}
	books, err := client.OrderBooks(ctx, []types.CLOBBookParams{{TokenID: "12345"}, {TokenID: "67890"}})
	if err != nil {
		t.Fatalf("OrderBooks: %v", err)
	}
	if len(books) != 2 || books[1].AssetID != "67890" {
		t.Fatalf("order books = %+v", books)
	}
	price, err := client.Price(ctx, "12345", "BUY")
	if err != nil {
		t.Fatalf("Price: %v", err)
	}
	if price != "0.45" {
		t.Fatalf("price = %q", price)
	}
	prices, err := client.Prices(ctx, []types.CLOBBookParams{{TokenID: "12345", Side: "BUY"}, {TokenID: "67890", Side: "SELL"}})
	if err != nil {
		t.Fatalf("Prices: %v", err)
	}
	if prices["12345"] != "0.45" || prices["67890"] != "0.55" {
		t.Fatalf("prices = %+v", prices)
	}
	midpoint, err := client.Midpoint(ctx, "12345")
	if err != nil {
		t.Fatalf("Midpoint: %v", err)
	}
	if midpoint != "0.50" {
		t.Fatalf("midpoint = %q", midpoint)
	}
	midpoints, err := client.Midpoints(ctx, []types.CLOBBookParams{{TokenID: "12345"}, {TokenID: "67890"}})
	if err != nil {
		t.Fatalf("Midpoints: %v", err)
	}
	if midpoints["12345"] != "0.50" || midpoints["67890"] != "0.52" {
		t.Fatalf("midpoints = %+v", midpoints)
	}
	spread, err := client.Spread(ctx, "12345")
	if err != nil {
		t.Fatalf("Spread: %v", err)
	}
	if spread != "0.02" {
		t.Fatalf("spread = %q", spread)
	}
	tickSize, err := client.TickSize(ctx, "12345")
	if err != nil {
		t.Fatalf("TickSize: %v", err)
	}
	if tickSize.MinimumTickSize != "0.01" || tickSize.MinimumOrderSize != "1" {
		t.Fatalf("tick size = %+v", tickSize)
	}
	negRisk, err := client.NegRisk(ctx, "12345")
	if err != nil {
		t.Fatalf("NegRisk: %v", err)
	}
	if negRisk.NegRisk {
		t.Fatalf("neg risk = %+v", negRisk)
	}
	feeRate, err := client.FeeRateBps(ctx, "12345")
	if err != nil {
		t.Fatalf("FeeRateBps: %v", err)
	}
	if feeRate != 30 {
		t.Fatalf("fee rate = %d", feeRate)
	}
	lastTrade, err := client.LastTradePrice(ctx, "12345")
	if err != nil {
		t.Fatalf("LastTradePrice: %v", err)
	}
	if lastTrade != "0.47" {
		t.Fatalf("last trade = %q", lastTrade)
	}
	lastTrades, err := client.LastTradesPrices(ctx, []types.CLOBBookParams{{TokenID: "12345"}, {TokenID: "67890"}})
	if err != nil {
		t.Fatalf("LastTradesPrices: %v", err)
	}
	if lastTrades["12345"] != "0.47" || lastTrades["67890"] != "0.53" {
		t.Fatalf("last trades = %+v", lastTrades)
	}
	history, err := client.PricesHistory(ctx, &types.CLOBPriceHistoryParams{Market: "12345", Interval: "1h", Fidelity: 60})
	if err != nil {
		t.Fatalf("PricesHistory: %v", err)
	}
	if len(history.History) != 1 || history.History[0].P != "0.45" {
		t.Fatalf("history = %+v", history)
	}
	scoring, err := client.OrderScoring(ctx, "0xorder")
	if err != nil {
		t.Fatalf("OrderScoring: %v", err)
	}
	if !scoring {
		t.Fatalf("order scoring = %v", scoring)
	}
	scoringRows, err := client.OrdersScoring(ctx, []string{"0xorder", "0xother"})
	if err != nil {
		t.Fatalf("OrdersScoring: %v", err)
	}
	if len(scoringRows) != 2 || !scoringRows[0] || scoringRows[1] {
		t.Fatalf("orders scoring = %+v", scoringRows)
	}
	if rewards, err := client.RewardsConfig(ctx); err != nil || len(rewards) != 1 || rewards[0].Market != "condition-1" {
		t.Fatalf("RewardsConfig = %+v err=%v", rewards, err)
	}
	if rawRewards, err := client.RawRewards(ctx, "condition-1"); err != nil || len(rawRewards) != 1 || rawRewards[0].RewardsPaid != 12 {
		t.Fatalf("RawRewards = %+v err=%v", rawRewards, err)
	}
	if earnings, err := client.UserEarnings(ctx, "2026-05-08"); err != nil || len(earnings) != 1 || earnings[0].Earnings != 3 {
		t.Fatalf("UserEarnings = %+v err=%v", earnings, err)
	}
	if totalEarnings, err := client.TotalEarnings(ctx, "2026-05-08"); err != nil || totalEarnings.Earnings != 9 {
		t.Fatalf("TotalEarnings = %+v err=%v", totalEarnings, err)
	}
	if rewardPercentages, err := client.RewardPercentages(ctx); err != nil || len(rewardPercentages) != 1 || rewardPercentages[0].RewardPercentage != 0.5 {
		t.Fatalf("RewardPercentages = %+v err=%v", rewardPercentages, err)
	}
	if userRewards, err := client.UserRewardsByMarket(ctx, nil); err != nil || len(userRewards) != 1 || userRewards[0].TotalRewards != 4 {
		t.Fatalf("UserRewardsByMarket = %+v err=%v", userRewards, err)
	}
	if rebates, err := client.RebatedFees(ctx); err != nil || len(rebates) != 1 || rebates[0].TotalRebated != 2 {
		t.Fatalf("RebatedFees = %+v err=%v", rebates, err)
	}

	orders, err := client.ListOrders(ctx, e2ePrivateKey)
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	if len(orders) != 1 || orders[0].ID != "0xorder" {
		t.Fatalf("orders = %+v", orders)
	}
	order, err := client.Order(ctx, e2ePrivateKey, "0xorder")
	if err != nil {
		t.Fatalf("Order: %v", err)
	}
	if order.OrderType != "GTC" {
		t.Fatalf("order = %+v", order)
	}
	clobTrades, err := client.ListTrades(ctx, e2ePrivateKey)
	if err != nil {
		t.Fatalf("ListTrades: %v", err)
	}
	if len(clobTrades) != 1 || clobTrades[0].TransactionHash != "0xtx" {
		t.Fatalf("clob trades = %+v", clobTrades)
	}
	balance, err := client.BalanceAllowance(ctx, e2ePrivateKey, clob.BalanceAllowanceParams{AssetType: "COLLATERAL"})
	if err != nil {
		t.Fatalf("BalanceAllowance: %v", err)
	}
	if balance.Balance != "1000000" || balance.Allowance != "999" {
		t.Fatalf("balance = %+v", balance)
	}
	updatedBalance, err := client.UpdateBalanceAllowance(ctx, e2ePrivateKey, clob.BalanceAllowanceParams{AssetType: "COLLATERAL"})
	if err != nil {
		t.Fatalf("UpdateBalanceAllowance: %v", err)
	}
	if updatedBalance.Balance != "1000000" {
		t.Fatalf("updated balance = %+v", updatedBalance)
	}
	cancelOne, err := client.CancelOrder(ctx, e2ePrivateKey, "0xorder")
	if err != nil {
		t.Fatalf("CancelOrder: %v", err)
	}
	if len(cancelOne.Canceled) != 1 || cancelOne.Canceled[0] != "0xorder" {
		t.Fatalf("cancel one = %+v", cancelOne)
	}
	cancelMany, err := client.CancelOrders(ctx, e2ePrivateKey, []string{"0xorder", "0xother"})
	if err != nil {
		t.Fatalf("CancelOrders: %v", err)
	}
	if len(cancelMany.Canceled) != 2 {
		t.Fatalf("cancel many = %+v", cancelMany)
	}
	cancelAll, err := client.CancelAll(ctx, e2ePrivateKey)
	if err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if len(cancelAll.Canceled) != 1 || cancelAll.Canceled[0] != "all" {
		t.Fatalf("cancel all = %+v", cancelAll)
	}
	cancelMarket, err := client.CancelMarket(ctx, e2ePrivateKey, clob.CancelMarketParams{Market: "condition-1", Asset: "12345"})
	if err != nil {
		t.Fatalf("CancelMarket: %v", err)
	}
	if len(cancelMarket.Canceled) != 1 || cancelMarket.Canceled[0] != "market" {
		t.Fatalf("cancel market = %+v", cancelMarket)
	}
	if err := client.Heartbeat(ctx, e2ePrivateKey, "hb-1"); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	limitOrder, err := client.CreateLimitOrder(ctx, e2ePrivateKey, clob.CreateOrderParams{
		TokenID:   "12345",
		Side:      "BUY",
		Price:     "0.500000",
		Size:      "2.000000",
		OrderType: "GTC",
		PostOnly:  true,
	})
	if err != nil {
		t.Fatalf("CreateLimitOrder: %v", err)
	}
	if limitOrder.OrderID != "0xposted" || limitOrder.Status != "live" {
		t.Fatalf("limit order = %+v", limitOrder)
	}
	batchOrder, err := client.CreateBatchOrders(ctx, e2ePrivateKey, []clob.CreateOrderParams{
		{TokenID: "12345", Side: "BUY", Price: "0.500000", Size: "1.000000", OrderType: "GTC"},
		{TokenID: "67890", Side: "SELL", Price: "0.600000", Size: "1.000000", OrderType: "GTC"},
	})
	if err != nil {
		t.Fatalf("CreateBatchOrders: %v", err)
	}
	if len(batchOrder.Orders) != 2 || batchOrder.Orders[1].OrderID != "0xposted-2" {
		t.Fatalf("batch order = %+v", batchOrder)
	}
	marketOrder, err := client.CreateMarketOrder(ctx, e2ePrivateKey, clob.MarketOrderParams{
		TokenID:   "12345",
		Side:      "BUY",
		Amount:    "1.000000",
		Price:     "0.500000",
		OrderType: "FOK",
	})
	if err != nil {
		t.Fatalf("CreateMarketOrder: %v", err)
	}
	if marketOrder.OrderID != "0xposted" {
		t.Fatalf("market order = %+v", marketOrder)
	}

	positions, err := client.CurrentPositionsWithLimit(ctx, "0xuser", 2)
	if err != nil {
		t.Fatalf("CurrentPositionsWithLimit: %v", err)
	}
	if len(positions) != 1 || positions[0].TokenID != "12345" {
		t.Fatalf("positions = %+v", positions)
	}
	closedPositions, err := client.ClosedPositionsWithLimit(ctx, "0xuser", 2)
	if err != nil {
		t.Fatalf("ClosedPositionsWithLimit: %v", err)
	}
	if len(closedPositions) != 1 || closedPositions[0].RealizedPnl != 1.5 {
		t.Fatalf("closed positions = %+v", closedPositions)
	}
	dataTrades, err := client.Trades(ctx, "0xuser", 2)
	if err != nil {
		t.Fatalf("Trades: %v", err)
	}
	if len(dataTrades) != 1 || dataTrades[0].ID != "data-trade-1" {
		t.Fatalf("data trades = %+v", dataTrades)
	}
	activity, err := client.Activity(ctx, "0xuser", 2)
	if err != nil {
		t.Fatalf("Activity: %v", err)
	}
	if len(activity) != 1 || activity[0].Type != "TRADE" {
		t.Fatalf("activity = %+v", activity)
	}
	holders, err := client.TopHolders(ctx, "condition-1", 2)
	if err != nil {
		t.Fatalf("TopHolders: %v", err)
	}
	if len(holders) != 1 || holders[0].Address != "0xholder" {
		t.Fatalf("holders = %+v", holders)
	}
	totalValue, err := client.TotalValue(ctx, "0xuser")
	if err != nil {
		t.Fatalf("TotalValue: %v", err)
	}
	if totalValue.Value != 42 {
		t.Fatalf("total value = %+v", totalValue)
	}
	marketsTraded, err := client.MarketsTraded(ctx, "0xuser")
	if err != nil {
		t.Fatalf("MarketsTraded: %v", err)
	}
	if marketsTraded.MarketsTraded != 3 {
		t.Fatalf("markets traded = %+v", marketsTraded)
	}
	openInterest, err := client.OpenInterest(ctx, "condition-1")
	if err != nil {
		t.Fatalf("OpenInterest: %v", err)
	}
	if openInterest.OpenValue != 25 {
		t.Fatalf("open interest = %+v", openInterest)
	}
	leaderboard, err := client.TraderLeaderboard(ctx, 2)
	if err != nil {
		t.Fatalf("TraderLeaderboard: %v", err)
	}
	if len(leaderboard) != 1 || leaderboard[0].Rank != 1 {
		t.Fatalf("leaderboard = %+v", leaderboard)
	}
	volume, err := client.LiveVolume(ctx, 2)
	if err != nil {
		t.Fatalf("LiveVolume: %v", err)
	}
	if volume.Total != 1 || volume.Events[0].EventID != "event-1" {
		t.Fatalf("live volume = %+v", volume)
	}

	signer, err := relayer.NewSigner(e2ePrivateKey, 137)
	if err != nil {
		t.Fatalf("relayer.NewSigner: %v", err)
	}
	relayerClient, err := relayer.NewV2(relayerServer.URL, relayer.V2APIKey{Key: "relayer-key", Address: signer.Address()}, 137)
	if err != nil {
		t.Fatalf("relayer.NewV2: %v", err)
	}
	deployed, err := relayerClient.IsDeployed(ctx, signer.Address())
	if err != nil {
		t.Fatalf("relayer IsDeployed: %v", err)
	}
	if !deployed {
		t.Fatal("relayer IsDeployed = false")
	}
	nonce, err := relayerClient.GetNonce(ctx, signer.Address())
	if err != nil {
		t.Fatalf("relayer GetNonce: %v", err)
	}
	if nonce != "7" {
		t.Fatalf("nonce = %q", nonce)
	}
	approvalCalls := relayer.BuildApprovalCalls()
	if len(approvalCalls) != 6 {
		t.Fatalf("approval calls = %d", len(approvalCalls))
	}
	deadline := relayer.BuildDeadline(60)
	signature, err := relayer.SignWalletBatch(signer, "0x1234567890123456789012345678901234567890", nonce, deadline, approvalCalls)
	if err != nil {
		t.Fatalf("relayer SignWalletBatch: %v", err)
	}
	if !strings.HasPrefix(signature, "0x") || len(signature) != 132 {
		t.Fatalf("signature len=%d value=%q", len(signature), signature)
	}

	if deriveCalls := rec.count("GET /auth/derive-api-key"); deriveCalls != 0 {
		t.Fatalf("configured CLOB credential flow called derive api key %d times", deriveCalls)
	}
	for _, route := range []string{
		"GET /markets", "GET /public-search", "GET /comments", "GET /clob-markets/condition-1",
		"GET /book", "POST /books", "POST /prices", "GET /data/orders", "GET /data/trades",
		"GET /balance-allowance", "DELETE /order", "DELETE /orders", "POST /order",
		"GET /positions", "GET /live-volume", "GET /deployed", "GET /nonce",
	} {
		if rec.count(route) == 0 {
			t.Fatalf("expected e2e route %s to be exercised", route)
		}
	}
}

func TestPolygolemSettlementE2EStopsAtRelayerAllowlistBlocker(t *testing.T) {
	rec := newE2ERecorder()
	dataServer := newE2ESettlementDataServer(t, rec)
	defer dataServer.Close()
	relayerServer := newE2EAllowlistRejectingRelayerServer(t, rec)
	defer relayerServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dataClient := data.NewClient(data.Config{BaseURL: dataServer.URL})
	positions, err := settlement.FindRedeemable(ctx, dataClient, "0x21999a074344610057c9b2B362332388a44502D4")
	if err != nil {
		t.Fatalf("FindRedeemable: %v", err)
	}
	if len(positions) != 2 {
		t.Fatalf("redeemable positions=%d want 2; got %+v", len(positions), positions)
	}
	if positions[0].NegativeRisk || !positions[1].NegativeRisk {
		t.Fatalf("negative-risk flags not preserved: %+v", positions)
	}

	signer, err := relayer.NewSigner(e2ePrivateKey, 137)
	if err != nil {
		t.Fatalf("relayer.NewSigner: %v", err)
	}
	relayerClient, err := relayer.NewV2(relayerServer.URL, relayer.V2APIKey{Key: "relayer-key", Address: signer.Address()}, 137)
	if err != nil {
		t.Fatalf("relayer.NewV2: %v", err)
	}
	result, err := settlement.SubmitRedeem(ctx, relayerClient, e2ePrivateKey, positions, settlement.DefaultBatchLimit)
	if err == nil {
		t.Fatalf("SubmitRedeem succeeded: %+v", result)
	}
	if result != nil {
		t.Fatalf("SubmitRedeem result=%+v want nil on rejected relay submit", result)
	}
	errText := err.Error()
	if !strings.Contains(errText, "upstream relayer allowlist blocker") {
		t.Fatalf("SubmitRedeem error=%q, want upstream blocker classification", errText)
	}
	if !strings.Contains(errText, "not in the allowed list") {
		t.Fatalf("SubmitRedeem error=%q, want relayer allowlist body preserved", errText)
	}
	if rec.count("GET /positions") != 1 {
		t.Fatalf("positions route count=%d want 1", rec.count("GET /positions"))
	}
	if rec.count("GET /nonce") != 1 {
		t.Fatalf("nonce route count=%d want 1", rec.count("GET /nonce"))
	}
	if rec.count("POST /submit") != 1 {
		t.Fatalf("submit route count=%d want 1", rec.count("POST /submit"))
	}
}

func TestPolygolemPublicSDKStreamE2EAgainstLocalWebSocket(t *testing.T) {
	upgrader := websocket.Upgrader{}
	subscriptions := make(chan map[string]interface{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(w, request, nil)
		if err != nil {
			t.Errorf("upgrade websocket: %v", err)
			return
		}
		defer conn.Close()

		var sub map[string]interface{}
		if err := conn.ReadJSON(&sub); err != nil {
			t.Errorf("read subscription: %v", err)
			return
		}
		subscriptions <- sub

		messages := []map[string]interface{}{
			{
				"event_type": "book",
				"asset_id":   "12345",
				"market":     "condition-1",
				"timestamp":  "1710000000",
				"hash":       "book-hash",
				"bids":       []map[string]string{{"price": "0.44", "size": "10"}},
				"asks":       []map[string]string{{"price": "0.46", "size": "11"}},
			},
			{
				"event_type": "price_change",
				"market":     "condition-1",
				"timestamp":  "1710000001",
				"price_changes": []map[string]string{{
					"asset_id": "12345",
					"price":    "0.45",
					"side":     "BUY",
					"size":     "20",
					"hash":     "change-hash",
					"best_bid": "0.45",
					"best_ask": "0.47",
				}},
			},
			{
				"event_type":       "last_trade_price",
				"asset_id":         "12345",
				"market":           "condition-1",
				"price":            "0.47",
				"side":             "BUY",
				"size":             "5",
				"fee_rate_bps":     "0",
				"timestamp":        "1710000002",
				"transaction_hash": "0xtx",
			},
		}
		for _, msg := range messages {
			if err := conn.WriteJSON(msg); err != nil {
				t.Errorf("write websocket message: %v", err)
				return
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := universal.NewClient(universal.Config{}).StreamClientWithConfig(stream.Config{
		URL:          wsURL,
		PingInterval: time.Hour,
		PongTimeout:  time.Second,
		Reconnect:    false,
	})
	bookCh := make(chan stream.BookMessage, 1)
	priceCh := make(chan stream.PriceChangeMessage, 1)
	tradeCh := make(chan stream.LastTradeMessage, 1)
	client.OnBook = func(msg stream.BookMessage) { bookCh <- msg }
	client.OnPriceChange = func(msg stream.PriceChangeMessage) { priceCh <- msg }
	client.OnLastTrade = func(msg stream.LastTradeMessage) { tradeCh <- msg }

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("stream Connect: %v", err)
	}
	defer client.Close()
	if err := client.SubscribeAssets(ctx, []string{"12345"}); err != nil {
		t.Fatalf("stream SubscribeAssets: %v", err)
	}
	if !client.IsConnected() {
		t.Fatal("stream client is not connected")
	}

	select {
	case sub := <-subscriptions:
		if sub["type"] != "market" {
			t.Fatalf("subscription type = %v", sub["type"])
		}
		assets, ok := sub["assets_ids"].([]interface{})
		if !ok || len(assets) != 1 || assets[0] != "12345" {
			t.Fatalf("subscription assets = %#v", sub["assets_ids"])
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for subscription")
	}
	select {
	case msg := <-bookCh:
		if msg.AssetID != "12345" || msg.Bids[0].Price != "0.44" {
			t.Fatalf("book message = %+v", msg)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for book message")
	}
	select {
	case msg := <-priceCh:
		if len(msg.PriceChanges) != 1 || msg.PriceChanges[0].BestAsk != "0.47" {
			t.Fatalf("price change message = %+v", msg)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for price change message")
	}
	select {
	case msg := <-tradeCh:
		if msg.TransactionHash != "0xtx" || msg.Price != "0.47" {
			t.Fatalf("last trade message = %+v", msg)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for last trade message")
	}

	dedup := stream.NewDeduplicator(10, time.Minute)
	rawMessage := []byte(`{"event_type":"book","hash":"same-message"}`)
	if !dedup.Process(rawMessage) {
		t.Fatal("deduplicator rejected first message")
	}
	if dedup.Process(rawMessage) {
		t.Fatal("deduplicator accepted duplicate message")
	}
	in, dup, out := dedup.Stats()
	if in != 2 || dup != 1 || out != 1 {
		t.Fatalf("deduplicator stats in=%d dup=%d out=%d", in, dup, out)
	}
}

type e2eRecorder struct {
	mu   sync.Mutex
	hits map[string]int
}

func newE2ERecorder() *e2eRecorder {
	return &e2eRecorder{hits: map[string]int{}}
}

func (r *e2eRecorder) hit(request *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hits[request.Method+" "+request.URL.Path]++
}

func (r *e2eRecorder) count(route string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.hits[route]
}

func newE2EGammaServer(t *testing.T, rec *e2eRecorder) *httptest.Server {
	t.Helper()
	category := map[string]interface{}{"id": "cat-1", "label": "Crypto", "slug": "crypto"}
	tag := map[string]interface{}{"id": "1", "label": "Crypto", "slug": "crypto"}
	market := map[string]interface{}{
		"id":                    "market-1",
		"question":              "Will BTC hit 100k?",
		"conditionId":           "condition-1",
		"slug":                  "will-btc-hit-100k",
		"category":              "Crypto",
		"categories":            []map[string]interface{}{category},
		"tags":                  []map[string]interface{}{tag},
		"active":                true,
		"closed":                false,
		"enableOrderBook":       true,
		"clobTokenIds":          `["12345","67890"]`,
		"orderPriceMinTickSize": 0.01,
		"orderMinSize":          5,
		"acceptingOrders":       true,
	}
	event := map[string]interface{}{
		"id":         "event-1",
		"slug":       "btc-event",
		"title":      "BTC event",
		"active":     true,
		"categories": []map[string]interface{}{category},
		"tags":       []map[string]interface{}{tag},
		"markets":    []map[string]interface{}{market},
	}
	series := map[string]interface{}{
		"id":         "series-1",
		"ticker":     "CRYPTO",
		"slug":       "crypto",
		"title":      "Crypto",
		"categories": []map[string]interface{}{category},
		"tags":       []map[string]interface{}{tag},
	}
	comment := map[string]interface{}{
		"id":   "comment-1",
		"body": "local sentiment",
		"user": map[string]interface{}{"address": "0xuser", "pseudonym": "local-user"},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		rec.hit(request)
		if request.Method != http.MethodGet {
			failRequest(t, w, "gamma unexpected method: %s %s", request.Method, request.URL.String())
			return
		}
		switch request.URL.Path {
		case "/":
			respondJSON(t, w, map[string]string{"data": "ok"})
		case "/markets":
			respondJSON(t, w, []map[string]interface{}{market})
		case "/markets/market-1", "/markets/will-btc-hit-100k":
			respondJSON(t, w, market)
		case "/events":
			respondJSON(t, w, []map[string]interface{}{event})
		case "/events/event-1":
			respondJSON(t, w, event)
		case "/public-search":
			respondJSON(t, w, map[string]interface{}{"events": []map[string]interface{}{event}})
		case "/series":
			respondJSON(t, w, []map[string]interface{}{series})
		case "/series/series-1":
			respondJSON(t, w, series)
		case "/tags":
			respondJSON(t, w, []map[string]interface{}{tag})
		case "/tags/1", "/tags/crypto":
			respondJSON(t, w, tag)
		case "/tags/1/related", "/tags/crypto/related":
			respondJSON(t, w, []map[string]interface{}{{"id": "rel-1", "tagID": 1, "relatedTagID": 2}})
		case "/teams":
			respondJSON(t, w, []map[string]interface{}{{"id": 1, "name": "Yankees", "league": "MLB"}})
		case "/sports-metadata":
			respondJSON(t, w, []map[string]interface{}{{"sport": "mlb", "image": "mlb.png", "resolution": "game"}})
		case "/comments":
			respondJSON(t, w, []map[string]interface{}{comment})
		case "/comments/comment-1":
			respondJSON(t, w, comment)
		case "/profiles/0xuser":
			respondJSON(t, w, map[string]interface{}{"id": "profile-1", "proxyWallet": "0xuser", "name": "Local User"})
		case "/sports-market-types":
			respondJSON(t, w, []map[string]interface{}{{"id": "moneyline", "name": "Moneyline", "slug": "moneyline"}})
		case "/markets/token/12345":
			respondJSON(t, w, map[string]interface{}{"market": market, "token_id": "12345", "outcome": "Yes"})
		case "/events-keyset":
			respondJSON(t, w, map[string]interface{}{"data": []map[string]interface{}{event}, "next_cursor": "event-cursor"})
		case "/markets-keyset":
			respondJSON(t, w, map[string]interface{}{"data": []map[string]interface{}{market}, "next_cursor": "market-cursor"})
		default:
			failRequest(t, w, "gamma unexpected route: %s %s", request.Method, request.URL.String())
		}
	}))
}

func newE2ECLOBServer(t *testing.T, rec *e2eRecorder) *httptest.Server {
	t.Helper()
	marketList := map[string]interface{}{
		"limit":       1,
		"count":       1,
		"next_cursor": "LTE=",
		"data": []map[string]interface{}{{
			"condition_id":              "condition-1",
			"question_id":               "question-1",
			"tokens":                    []map[string]interface{}{{"token_id": "12345", "outcome": "Yes", "price": "0.45"}, {"token_id": "67890", "outcome": "No", "price": "0.55"}},
			"enable_order_book":         true,
			"order_price_min_tick_size": 0.01,
			"order_min_size":            5,
			"accepting_orders":          true,
		}},
	}
	book := func(tokenID string) map[string]interface{} {
		return map[string]interface{}{
			"market":           "condition-1",
			"asset_id":         tokenID,
			"timestamp":        "1710000000",
			"hash":             "book-hash-" + tokenID,
			"bids":             []map[string]string{{"price": "0.44", "size": "10"}},
			"asks":             []map[string]string{{"price": "0.46", "size": "11"}},
			"min_order_size":   "1",
			"tick_size":        "0.01",
			"last_trade_price": "0.47",
		}
	}
	cancelResponse := map[string]interface{}{"canceled": []string{"0xorder"}, "not_canceled": map[string]string{}}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		rec.hit(request)
		switch request.URL.Path {
		case "/":
			respondJSON(t, w, map[string]string{"data": "ok"})
		case "/auth/api-key":
			if request.Method != http.MethodPost {
				failRequest(t, w, "auth api-key method = %s", request.Method)
				return
			}
			if request.Header.Get("POLY_ADDRESS") == "" || request.Header.Get("POLY_SIGNATURE") == "" {
				failRequest(t, w, "auth api-key missing L1 headers")
				return
			}
			respondJSON(t, w, map[string]string{"apiKey": "created-key", "secret": e2eConfiguredCLOBSecret, "passphrase": "created-pass"})
		case "/auth/derive-api-key":
			failRequest(t, w, "configured e2e flow must not derive api key")
		case "/time":
			respondJSON(t, w, map[string]string{"timestamp": "1710000000", "iso": "2026-05-08T00:00:00Z"})
		case "/markets", "/simplified-markets", "/sampling-markets", "/sampling-simplified-markets":
			respondJSON(t, w, marketList)
		case "/clob-markets/condition-1":
			respondJSON(t, w, map[string]interface{}{
				"gst":   "2026-05-08T00:00:00Z",
				"t":     []map[string]string{{"t": "12345", "o": "Yes"}, {"t": "67890", "o": "No"}},
				"mos":   5,
				"mts":   0.01,
				"mbf":   0,
				"tbf":   0,
				"rfqe":  true,
				"itode": true,
				"ibce":  true,
				"fd":    map[string]interface{}{"r": 0.02, "e": 2, "to": true},
				"oas":   5,
			})
		case "/book":
			respondJSON(t, w, book(request.URL.Query().Get("token_id")))
		case "/books":
			if request.Method != http.MethodPost {
				failRequest(t, w, "books method = %s", request.Method)
				return
			}
			respondJSON(t, w, []map[string]interface{}{book("12345"), book("67890")})
		case "/price":
			respondJSON(t, w, map[string]interface{}{"price": 0.45})
		case "/prices":
			respondJSON(t, w, map[string]interface{}{"12345": map[string]interface{}{"BUY": 0.45}, "67890": map[string]interface{}{"SELL": "0.55"}})
		case "/midpoint":
			respondJSON(t, w, map[string]interface{}{"mid_price": "0.50"})
		case "/midpoints":
			respondJSON(t, w, map[string]interface{}{"12345": "0.50", "67890": "0.52"})
		case "/spread":
			respondJSON(t, w, map[string]string{"spread": "0.02"})
		case "/tick-size":
			respondJSON(t, w, map[string]interface{}{"minimum_tick_size": "0.01", "minimum_order_size": "1", "tick_size": "0.01"})
		case "/neg-risk":
			respondJSON(t, w, map[string]interface{}{"neg_risk": false, "neg_risk_market_id": "neg-1", "neg_risk_fee_bips": 0})
		case "/fee-rate":
			respondJSON(t, w, map[string]int{"base_fee": 30})
		case "/last-trade-price":
			respondJSON(t, w, map[string]interface{}{"price": 0.47})
		case "/last-trades-prices":
			respondJSON(t, w, []map[string]interface{}{{"token_id": "12345", "price": "0.47"}, {"token_id": "67890", "price": "0.53"}})
		case "/prices-history":
			respondJSON(t, w, map[string]interface{}{"history": []map[string]interface{}{{"t": 1710000000, "p": 0.45, "v": 10}}})
		case "/orders/scoring":
			if request.Method == http.MethodGet {
				respondJSON(t, w, map[string]bool{"scoring": true})
				return
			}
			respondJSON(t, w, []bool{true, false})
		case "/rewards/config":
			respondJSON(t, w, []map[string]interface{}{{"market": "condition-1", "asset_address": "12345", "rewards_min_size": 1, "rewards_max_spread": 0.03, "active": true}})
		case "/rewards/raw":
			respondJSON(t, w, []map[string]interface{}{{"market": "condition-1", "date": "2026-05-08", "rewards_paid": 12, "volume": 100}})
		case "/rewards/earnings":
			respondJSON(t, w, []map[string]interface{}{{"date": "2026-05-08", "earnings": 3, "market": "condition-1"}})
		case "/rewards/total-earnings":
			respondJSON(t, w, map[string]interface{}{"date": "2026-05-08", "earnings": 9})
		case "/rewards/percentages":
			respondJSON(t, w, []map[string]interface{}{{"market": "condition-1", "reward_percentage": 0.5}})
		case "/rewards/markets":
			respondJSON(t, w, []map[string]interface{}{{"market": "condition-1", "total_rewards": 4, "reward_percentage": 0.5}})
		case "/rebates":
			respondJSON(t, w, []map[string]interface{}{{"maker_address": "0xmaker", "market": "condition-1", "total_rebated": 2, "date": "2026-05-08"}})
		case "/data/orders":
			expectConfiguredCLOBAuth(t, w, request)
			respondJSON(t, w, map[string]interface{}{"orders": []map[string]interface{}{{
				"id": "0xorder", "status": "ORDER_STATUS_LIVE", "market": "condition-1", "asset_id": "12345", "side": "BUY",
				"original_size": "10", "size_matched": "2", "price": "0.45", "outcome": "Yes", "order_type": "GTC",
				"maker_address": "0xmaker", "owner": e2eConfiguredCLOBKey, "created_at": "1710000000", "expiration": "0",
			}}, "next_cursor": "LTE=", "count": 1})
		case "/data/trades":
			expectConfiguredCLOBAuth(t, w, request)
			respondJSON(t, w, map[string]interface{}{"trades": []map[string]interface{}{{
				"id": "trade-1", "status": "MATCHED", "market": "condition-1", "asset_id": "12345", "side": "BUY",
				"price": "0.45", "size": "2", "fee_rate_bps": "0", "outcome": "Yes", "owner": e2eConfiguredCLOBKey,
				"builder": "builder", "matched_amount": "2", "transaction_hash": "0xtx", "created_at": "1710000000", "last_updated": "1710000001",
			}}, "next_cursor": "LTE=", "count": 1})
		case "/order/0xorder":
			expectConfiguredCLOBAuth(t, w, request)
			respondJSON(t, w, map[string]interface{}{"id": "0xorder", "status": "ORDER_STATUS_LIVE", "order_type": "GTC"})
		case "/balance-allowance", "/balance-allowance/update":
			expectConfiguredCLOBAuth(t, w, request)
			if request.URL.Query().Get("signature_type") != "3" || request.URL.Query().Get("asset_type") != "COLLATERAL" {
				failRequest(t, w, "unexpected balance-allowance query: %s", request.URL.RawQuery)
				return
			}
			respondJSON(t, w, map[string]interface{}{"balance": "1000000", "allowance": "999", "allowances": map[string]string{"exchange": "999"}})
		case "/order":
			expectConfiguredCLOBAuth(t, w, request)
			if request.Method == http.MethodDelete {
				respondJSON(t, w, cancelResponse)
				return
			}
			var body map[string]interface{}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				failRequest(t, w, "decode order body: %v", err)
				return
			}
			if body["owner"] != e2eConfiguredCLOBKey {
				failRequest(t, w, "order owner = %#v", body["owner"])
				return
			}
			order, ok := body["order"].(map[string]interface{})
			if !ok || order["signatureType"] != float64(3) || order["builder"] != e2eBuilderCode {
				failRequest(t, w, "unexpected signed order body: %#v", body)
				return
			}
			respondJSON(t, w, map[string]interface{}{"success": true, "orderID": "0xposted", "status": "live"})
		case "/orders":
			expectConfiguredCLOBAuth(t, w, request)
			if request.Method == http.MethodDelete {
				respondJSON(t, w, map[string]interface{}{"canceled": []string{"0xorder", "0xother"}, "not_canceled": map[string]string{}})
				return
			}
			var body []map[string]interface{}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				failRequest(t, w, "decode orders body: %v", err)
				return
			}
			if len(body) != 2 {
				failRequest(t, w, "batch order count = %d", len(body))
				return
			}
			respondJSON(t, w, []map[string]interface{}{{"success": true, "orderID": "0xposted-1", "status": "live"}, {"success": true, "orderID": "0xposted-2", "status": "live"}})
		case "/cancel-all":
			expectConfiguredCLOBAuth(t, w, request)
			respondJSON(t, w, map[string]interface{}{"canceled": []string{"all"}, "not_canceled": map[string]string{}})
		case "/cancel-market-orders":
			expectConfiguredCLOBAuth(t, w, request)
			respondJSON(t, w, map[string]interface{}{"canceled": []string{"market"}, "not_canceled": map[string]string{}})
		case "/v1/heartbeats":
			expectConfiguredCLOBAuth(t, w, request)
			respondJSON(t, w, map[string]string{"ok": "true"})
		default:
			failRequest(t, w, "clob unexpected route: %s %s", request.Method, request.URL.String())
		}
	}))
}

func newE2EDataServer(t *testing.T, rec *e2eRecorder) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		rec.hit(request)
		if request.Method != http.MethodGet {
			failRequest(t, w, "data unexpected method: %s %s", request.Method, request.URL.String())
			return
		}
		switch request.URL.Path {
		case "/":
			respondJSON(t, w, map[string]string{"data": "ok"})
		case "/positions":
			respondJSON(t, w, []map[string]interface{}{{"asset": "12345", "conditionId": "condition-1", "eventId": "event-1", "size": 2, "avgPrice": 0.4, "curPrice": 0.45, "cashPnl": 0.1, "redeemable": false}})
		case "/closed-positions":
			respondJSON(t, w, []map[string]interface{}{{"token_id": "12345", "condition_id": "condition-1", "market_id": "market-1", "side": "BUY", "avg_price_buy": 0.4, "avg_price_sell": 0.55, "size": 2, "realized_pnl": 1.5}})
		case "/trades":
			respondJSON(t, w, []map[string]interface{}{{"id": "data-trade-1", "market": "condition-1", "asset_id": "12345", "side": "BUY", "price": 0.45, "size": 2, "fee_rate_bps": 0, "created_at": "1710000000"}})
		case "/activity":
			respondJSON(t, w, []map[string]interface{}{{"type": "TRADE", "market": "condition-1", "asset_id": "12345", "side": "BUY", "price": 0.45, "size": 2, "timestamp": 1710000000}})
		case "/holders":
			respondJSON(t, w, []map[string]interface{}{{"token": "12345", "holders": []map[string]interface{}{{"proxyWallet": "0xholder", "amount": 3}}}})
		case "/value":
			respondJSON(t, w, []map[string]interface{}{{"user": "0xuser", "value": 42}})
		case "/traded":
			respondJSON(t, w, map[string]interface{}{"user": "0xuser", "traded": 3})
		case "/oi":
			respondJSON(t, w, []map[string]interface{}{{"market": "condition-1", "value": 25}})
		case "/v1/leaderboard":
			respondJSON(t, w, []map[string]interface{}{{"rank": "1", "proxyWallet": "0xuser", "vol": 100, "pnl": 5}})
		case "/live-volume":
			respondJSON(t, w, map[string]interface{}{"total": 1, "events": []map[string]interface{}{{"event_id": "event-1", "event_slug": "btc-event", "title": "BTC event", "volume": 1000}}})
		default:
			failRequest(t, w, "data unexpected route: %s %s", request.Method, request.URL.String())
		}
	}))
}

func newE2ERelayerServer(t *testing.T, rec *e2eRecorder) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		rec.hit(request)
		if request.Header.Get("RELAYER_API_KEY") != "relayer-key" || request.Header.Get("RELAYER_API_KEY_ADDRESS") == "" {
			failRequest(t, w, "missing relayer v2 headers")
			return
		}
		switch request.URL.Path {
		case "/deployed":
			respondJSON(t, w, map[string]interface{}{"deployed": true, "address": request.URL.Query().Get("address")})
		case "/nonce":
			respondJSON(t, w, map[string]string{"nonce": "7"})
		default:
			failRequest(t, w, "relayer unexpected route: %s %s", request.Method, request.URL.String())
		}
	}))
}

func newE2ESettlementDataServer(t *testing.T, rec *e2eRecorder) *httptest.Server {
	t.Helper()
	standardCondition := "0x1111111111111111111111111111111111111111111111111111111111111111"
	negRiskCondition := "0x2222222222222222222222222222222222222222222222222222222222222222"
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		rec.hit(request)
		if request.Method != http.MethodGet {
			failRequest(t, w, "settlement data unexpected method: %s %s", request.Method, request.URL.String())
			return
		}
		switch request.URL.Path {
		case "/positions":
			respondJSON(t, w, []map[string]interface{}{
				{
					"asset":       "non-redeemable-token",
					"conditionId": standardCondition,
					"redeemable":  false,
					"size":        3.0,
					"outcome":     "No",
					"title":       "Ignore losing/non-redeemable position",
				},
				{
					"asset":        "standard-winning-token",
					"conditionId":  standardCondition,
					"redeemable":   true,
					"negativeRisk": false,
					"size":         2.86,
					"outcome":      "Yes",
					"title":        "Standard winner",
					"slug":         "standard-winner",
				},
				{
					"asset":        "neg-risk-winning-token",
					"conditionId":  negRiskCondition,
					"redeemable":   true,
					"negativeRisk": true,
					"size":         1.25,
					"outcome":      "Yes",
					"title":        "Negative-risk winner",
					"slug":         "neg-risk-winner",
				},
			})
		default:
			failRequest(t, w, "settlement data unexpected route: %s %s", request.Method, request.URL.String())
		}
	}))
}

func newE2EAllowlistRejectingRelayerServer(t *testing.T, rec *e2eRecorder) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		rec.hit(request)
		if request.Header.Get("RELAYER_API_KEY") != "relayer-key" || request.Header.Get("RELAYER_API_KEY_ADDRESS") == "" {
			failRequest(t, w, "missing relayer v2 headers")
			return
		}
		switch request.URL.Path {
		case "/nonce":
			respondJSON(t, w, map[string]string{"nonce": "11"})
		case "/submit":
			if request.Method != http.MethodPost {
				failRequest(t, w, "relayer submit method=%s", request.Method)
				return
			}
			var body struct {
				Type                string `json:"type"`
				To                  string `json:"to"`
				DepositWalletParams struct {
					Calls []struct {
						Target string `json:"target"`
						Value  string `json:"value"`
						Data   string `json:"data"`
					} `json:"calls"`
				} `json:"depositWalletParams"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				failRequest(t, w, "decode relayer submit body: %v", err)
				return
			}
			if body.Type != "WALLET" {
				failRequest(t, w, "relayer submit type=%q want WALLET", body.Type)
				return
			}
			if !strings.EqualFold(body.To, contracts.DepositWalletFactory) {
				failRequest(t, w, "relayer submit to=%q want factory %s", body.To, contracts.DepositWalletFactory)
				return
			}
			if len(body.DepositWalletParams.Calls) != 2 {
				failRequest(t, w, "redeem call count=%d want 2", len(body.DepositWalletParams.Calls))
				return
			}
			seenTargets := map[string]bool{}
			for i, call := range body.DepositWalletParams.Calls {
				if strings.EqualFold(call.Target, contracts.CTF) {
					failRequest(t, w, "call %d targets raw CTF %s", i, call.Target)
					return
				}
				if call.Value != "0" {
					failRequest(t, w, "call %d value=%s want 0", i, call.Value)
					return
				}
				if !strings.HasPrefix(call.Data, "0x01b7037c") {
					failRequest(t, w, "call %d data selector=%q want redeemPositions", i, call.Data)
					return
				}
				seenTargets[strings.ToLower(call.Target)] = true
			}
			if !seenTargets[strings.ToLower(contracts.CtfCollateralAdapter)] {
				failRequest(t, w, "missing standard collateral adapter target")
				return
			}
			if !seenTargets[strings.ToLower(contracts.NegRiskCtfCollateralAdapter)] {
				failRequest(t, w, "missing negative-risk collateral adapter target")
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			respondJSON(t, w, map[string]interface{}{"error": "not in the allowed list", "code": 400})
		default:
			failRequest(t, w, "relayer unexpected route: %s %s", request.Method, request.URL.String())
		}
	}))
}

func expectConfiguredCLOBAuth(t *testing.T, w http.ResponseWriter, request *http.Request) {
	t.Helper()
	if request.Header.Get("POLY_API_KEY") != e2eConfiguredCLOBKey {
		failRequest(t, w, "POLY_API_KEY = %q", request.Header.Get("POLY_API_KEY"))
		return
	}
	for _, header := range []string{"POLY_ADDRESS", "POLY_TIMESTAMP", "POLY_SIGNATURE", "POLY_PASSPHRASE"} {
		if request.Header.Get(header) == "" {
			failRequest(t, w, "missing CLOB auth header %s", header)
			return
		}
	}
}

func respondJSON(t *testing.T, w http.ResponseWriter, body interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Errorf("write response: %v", err)
	}
}

func failRequest(t *testing.T, w http.ResponseWriter, format string, args ...interface{}) {
	t.Helper()
	t.Errorf(format, args...)
	http.Error(w, "unexpected e2e request", http.StatusBadRequest)
}
