package orderresults

import (
	"context"
	"testing"

	sdkclob "github.com/TrebuchetDynamics/polygolem/pkg/clob"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

func TestBuildReportJoinsPositionsClosedPositionsAndTrades(t *testing.T) {
	source := &fakeSource{
		positions: []types.Position{
			{
				TokenID:      "token-sol-up",
				ConditionID:  "0xsol",
				Size:         2.8627,
				AvgPrice:     0.5099,
				InitialValue: 1.46,
				CurrentValue: 2.8627,
				CurrentPrice: 1,
				CashPnl:      1.4027,
				Redeemable:   true,
				Outcome:      "Up",
				Title:        "Solana Up or Down - May 9, 8:20AM-8:25AM ET",
				Slug:         "sol-updown-5m-1778329200",
			},
			{
				TokenID:      "token-eth-up",
				ConditionID:  "0xeth",
				Size:         4.0784,
				AvgPrice:     0.5099,
				InitialValue: 2.08,
				CurrentValue: 0,
				CurrentPrice: 0,
				CashPnl:      -2.08,
				Redeemable:   true,
				Outcome:      "Up",
				Title:        "Ethereum Up or Down - May 9, 4:40AM-4:45AM ET",
				Slug:         "eth-updown-5m-1778316000",
			},
		},
		closed: []types.ClosedPosition{
			{},
			{
				TokenID:      "token-closed",
				ConditionID:  "0xclosed",
				Size:         5,
				AvgPrice:     0.42,
				RealizedPnl:  1.25,
				CurrentPrice: 1,
				Title:        "Closed market",
				Slug:         "closed-market",
				Outcome:      "Yes",
			},
		},
		trades: []types.Trade{
			{
				ID:              "trade-sol",
				Market:          "0xsol",
				AssetID:         "token-sol-up",
				Side:            "BUY",
				Price:           0.51,
				Size:            2.8627,
				Outcome:         "Up",
				TransactionHash: "0xsoltx",
				CreatedAt:       "1778329200",
			},
			{
				ID:        "trade-eth",
				Market:    "0xeth",
				AssetID:   "token-eth-up",
				Side:      "BUY",
				Price:     0.51,
				Size:      4.0784,
				Outcome:   "Up",
				CreatedAt: "1778316000",
			},
		},
	}

	report, err := BuildReport(context.Background(), source, "0xwallet", Options{Limit: 50})
	if err != nil {
		t.Fatal(err)
	}

	if report.User != "0xwallet" {
		t.Fatalf("user=%q", report.User)
	}
	if report.Summary.Won != 1 || report.Summary.Lost != 1 || report.Summary.Closed != 1 || report.Summary.Redeemable != 2 {
		t.Fatalf("summary=%+v", report.Summary)
	}
	if report.Summary.DataTrades != 2 {
		t.Fatalf("data trades=%d", report.Summary.DataTrades)
	}
	if got := len(report.Rows); got != 3 {
		t.Fatalf("rows=%d want 3: %+v", got, report.Rows)
	}

	sol := report.RowByToken("token-sol-up")
	if sol == nil {
		t.Fatal("missing sol row")
	}
	if sol.Status != StatusWon || !sol.Redeemable {
		t.Fatalf("sol status=%q redeemable=%v", sol.Status, sol.Redeemable)
	}
	if sol.TradeCount != 1 || sol.Trades[0].TransactionHash != "0xsoltx" {
		t.Fatalf("sol trades=%+v", sol.Trades)
	}

	eth := report.RowByToken("token-eth-up")
	if eth == nil {
		t.Fatal("missing eth row")
	}
	if eth.Status != StatusLost {
		t.Fatalf("eth status=%q", eth.Status)
	}

	closed := report.RowByToken("token-closed")
	if closed == nil {
		t.Fatal("missing closed row")
	}
	if closed.Status != StatusClosed || closed.RealizedPnl != 1.25 {
		t.Fatalf("closed=%+v", closed)
	}
}

func TestBuildReportCanIncludeAuthenticatedCLOBHistory(t *testing.T) {
	source := &fakeSource{
		clobOrders: []sdkclob.OrderRecord{{
			ID:           "order-live",
			Status:       "ORDER_STATUS_LIVE",
			Market:       "0xopen",
			AssetID:      "token-open",
			Side:         "BUY",
			Price:        "0.48",
			OriginalSize: "5",
			SizeMatched:  "0",
			Outcome:      "Up",
			CreatedAt:    "1778329500",
		}},
		clobTrades: []sdkclob.TradeRecord{{
			ID:              "clob-trade",
			Status:          "TRADE_STATUS_CONFIRMED",
			Market:          "0xopen",
			AssetID:         "token-open",
			Side:            "BUY",
			Price:           "0.48",
			Size:            "5",
			Outcome:         "Up",
			TransactionHash: "0xclobtx",
			CreatedAt:       "1778329501",
		}},
	}

	report, err := BuildReport(context.Background(), source, "0xwallet", Options{
		Limit:       20,
		IncludeCLOB: true,
		PrivateKey:  "0xprivate",
	})
	if err != nil {
		t.Fatal(err)
	}

	if source.seenPrivateKey != "0xprivate" {
		t.Fatalf("private key not passed to clob source")
	}
	row := report.RowByToken("token-open")
	if row == nil {
		t.Fatal("missing clob-only row")
	}
	if row.Status != StatusOpen || row.OpenOrderCount != 1 || row.TradeCount != 1 {
		t.Fatalf("row=%+v", row)
	}
	if row.OpenOrders[0].ID != "order-live" || row.Trades[0].Source != SourceCLOB {
		t.Fatalf("row details=%+v", row)
	}
}

func TestBuildReportDeduplicatesDataAndCLOBTradesByTransactionHash(t *testing.T) {
	source := &fakeSource{
		trades: []types.Trade{{
			Market:          "0xsol",
			AssetID:         "token-sol-up",
			Side:            "BUY",
			Price:           0.51,
			Size:            2,
			Outcome:         "Up",
			TransactionHash: "0xdupe",
		}},
		clobTrades: []sdkclob.TradeRecord{{
			ID:              "clob-trade",
			Status:          "CONFIRMED",
			Market:          "0xsol",
			AssetID:         "token-sol-up",
			Side:            "BUY",
			Price:           "0.51",
			Size:            "2",
			Outcome:         "Up",
			TransactionHash: "0xdupe",
		}},
	}

	report, err := BuildReport(context.Background(), source, "0xwallet", Options{
		IncludeCLOB: true,
		PrivateKey:  "0xprivate",
	})
	if err != nil {
		t.Fatal(err)
	}
	row := report.RowByToken("token-sol-up")
	if row == nil {
		t.Fatal("missing row")
	}
	if row.TradeCount != 1 || row.Trades[0].Source != SourceCLOB || row.Trades[0].ID != "clob-trade" {
		t.Fatalf("trades not deduplicated with CLOB preferred: %+v", row.Trades)
	}
	if report.Summary.MatchedNotional != 1.02 {
		t.Fatalf("matched notional=%v want 1.02", report.Summary.MatchedNotional)
	}
}

type fakeSource struct {
	positions      []types.Position
	closed         []types.ClosedPosition
	trades         []types.Trade
	clobOrders     []sdkclob.OrderRecord
	clobTrades     []sdkclob.TradeRecord
	seenPrivateKey string
}

func (f *fakeSource) CurrentPositionsWithLimit(context.Context, string, int) ([]types.Position, error) {
	return f.positions, nil
}

func (f *fakeSource) ClosedPositionsWithLimit(context.Context, string, int) ([]types.ClosedPosition, error) {
	return f.closed, nil
}

func (f *fakeSource) Trades(context.Context, string, int) ([]types.Trade, error) {
	return f.trades, nil
}

func (f *fakeSource) ListOrders(_ context.Context, privateKey string) ([]sdkclob.OrderRecord, error) {
	f.seenPrivateKey = privateKey
	return f.clobOrders, nil
}

func (f *fakeSource) ListTrades(_ context.Context, privateKey string) ([]sdkclob.TradeRecord, error) {
	f.seenPrivateKey = privateKey
	return f.clobTrades, nil
}
