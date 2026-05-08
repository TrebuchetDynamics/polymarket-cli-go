package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestVersionCommandPrintsVersion(t *testing.T) {
	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "test-version") {
		t.Fatalf("version output %q does not include test-version", got)
	}
}

func TestJSONFlagIsAcceptedAndPreflightEmitsJSON(t *testing.T) {
	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{"--json", "preflight"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var got struct {
		OK     bool `json:"ok"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("preflight stdout is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	if len(got.Checks) == 0 {
		t.Fatalf("preflight JSON checks is empty: %s", stdout.String())
	}
}

func TestDocumentedSubcommandsAreRegistered(t *testing.T) {
	for _, args := range [][]string{
		{"discover", "search"},
		{"discover", "markets"},
		{"discover", "market"},
		{"discover", "enrich"},
		{"discover", "tags"},
		{"discover", "series"},
		{"discover", "comments"},
		{"orderbook", "get"},
		{"orderbook", "price"},
		{"orderbook", "midpoint"},
		{"orderbook", "spread"},
		{"orderbook", "tick-size"},
		{"orderbook", "fee-rate"},
		{"orderbook", "last-trade"},
		{"clob", "book"},
		{"clob", "tick-size"},
		{"clob", "create-api-key"},
		{"clob", "balance"},
		{"clob", "update-balance"},
		{"clob", "orders"},
		{"clob", "order"},
		{"clob", "trades"},
		{"clob", "cancel"},
		{"clob", "cancel-orders"},
		{"clob", "cancel-all"},
		{"clob", "cancel-market"},
		{"clob", "create-order"},
		{"clob", "market-order"},
		{"clob", "price-history"},
		{"clob", "market"},
		{"clob", "markets"},
		{"data", "positions"},
		{"data", "closed-positions"},
		{"data", "trades"},
		{"data", "activity"},
		{"data", "holders"},
		{"data", "value"},
		{"data", "markets-traded"},
		{"data", "open-interest"},
		{"data", "leaderboard"},
		{"data", "live-volume"},
		{"stream", "market"},
		{"events", "list"},
		{"bridge", "assets"},
		{"bridge", "deposit"},
		{"health"},
		{"paper", "buy"},
		{"paper", "sell"},
		{"paper", "positions"},
		{"paper", "reset"},
		{"auth", "status"},
		{"live", "status"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
			root.SetArgs(append(args, "--help"))
			if err := root.Execute(); err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			wantPath := "polygolem " + strings.Join(args, " ")
			if !strings.Contains(stdout.String(), wantPath) {
				t.Fatalf("help output does not identify exact command path %q:\n%s", wantPath, stdout.String())
			}
		})
	}
}

func TestCLOBSignatureTypeDefaultsToDepositWallet(t *testing.T) {
	for _, args := range [][]string{
		{"clob", "balance"},
		{"clob", "update-balance"},
		{"clob", "create-order"},
		{"clob", "market-order"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			root := NewRootCommand(Options{Version: "test-version", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
			cmd, _, err := root.Find(args)
			if err != nil {
				t.Fatalf("Find returned error: %v", err)
			}
			flag := cmd.Flags().Lookup("signature-type")
			if flag == nil {
				t.Fatalf("signature-type flag missing")
			}
			if flag.DefValue != "deposit" {
				t.Fatalf("default signature-type=%q, want deposit", flag.DefValue)
			}
		})
	}
}

func TestStreamMarketReadsFromLocalWebSocket(t *testing.T) {
	upgrader := websocket.Upgrader{}
	subscriptions := make(chan []string, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade websocket: %v", err)
			return
		}
		defer conn.Close()

		var sub struct {
			Type     string   `json:"type"`
			AssetIDs []string `json:"assets_ids"`
		}
		if err := conn.ReadJSON(&sub); err != nil {
			t.Errorf("read subscription: %v", err)
			return
		}
		subscriptions <- sub.AssetIDs
		if err := conn.WriteJSON(map[string]any{
			"event_type": "book",
			"asset_id":   "token-1",
			"market":     "market-1",
			"timestamp":  "1",
			"bids":       []map[string]string{{"price": "0.50", "size": "10"}},
			"asks":       []map[string]string{{"price": "0.51", "size": "12"}},
		}); err != nil {
			t.Errorf("write stream message: %v", err)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{
		"--json",
		"stream", "market",
		"--url", "ws" + strings.TrimPrefix(server.URL, "http"),
		"--asset-ids", "token-1",
		"--max-messages", "1",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	gotSubscription := <-subscriptions
	if len(gotSubscription) != 1 || gotSubscription[0] != "token-1" {
		t.Fatalf("subscription=%v, want [token-1]", gotSubscription)
	}
	var got struct {
		EventType string `json:"event_type"`
		AssetID   string `json:"asset_id"`
		Market    string `json:"market"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stream stdout is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	if got.EventType != "book" || got.AssetID != "token-1" || got.Market != "market-1" {
		t.Fatalf("unexpected stream output: %+v", got)
	}
}

func TestDocumentedSubcommandArgsAreNotHandledByParentOnly(t *testing.T) {
	var stdout bytes.Buffer
	root := NewRootCommand(Options{Version: "test-version", Stdout: &stdout, Stderr: &bytes.Buffer{}})
	root.SetArgs([]string{"discover", "search", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(stdout.String(), "polygolem discover search") {
		t.Fatalf("discover search was not handled by its own command:\n%s", stdout.String())
	}
}

func TestNormalizeCollateralBalanceResponseScalesBaseUnits(t *testing.T) {
	raw := map[string]interface{}{
		"balance": "14000000",
		"allowances": map[string]string{
			"0xspender": "1000000",
		},
	}

	got := normalizeCollateralBalanceResponse(raw)

	if got["balance"] != "14.000000" {
		t.Fatalf("balance=%v", got["balance"])
	}
	if !reflect.DeepEqual(got["allowances"], raw["allowances"]) {
		t.Fatalf("allowances changed: %#v", got["allowances"])
	}
}
