package orderbook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReaderOrderBookSortsBestLevelsFirst(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/book" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if got := r.URL.Query().Get("token_id"); got != "token-1" {
			t.Fatalf("token_id=%s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"market": "condition-1",
			"asset_id": "token-1",
			"bids": [
				{"price": "0.01", "size": "100"},
				{"price": "0.25", "size": "5"}
			],
			"asks": [
				{"price": "0.99", "size": "100"},
				{"price": "0.26", "size": "5"}
			]
		}`))
	}))
	defer server.Close()

	reader := NewReader(server.URL)
	book, err := reader.OrderBook(context.Background(), "token-1")
	if err != nil {
		t.Fatal(err)
	}

	if len(book.Bids) != 2 || book.Bids[0].Price != 0.25 {
		t.Fatalf("bids not best-first: %#v", book.Bids)
	}
	if len(book.Asks) != 2 || book.Asks[0].Price != 0.26 {
		t.Fatalf("asks not best-first: %#v", book.Asks)
	}
}

func TestReaderOrderBooksPostsBatchTokens(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s, want POST", r.Method)
		}
		if r.URL.Path != "/books" {
			t.Fatalf("path=%s, want /books", r.URL.Path)
		}
		var body []struct {
			TokenID string `json:"token_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(body) != 2 || body[0].TokenID != "up-token" || body[1].TokenID != "down-token" {
			t.Fatalf("body=%#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"market": "condition-1",
				"asset_id": "up-token",
				"bids": [{"price": "0.44", "size": "10"}],
				"asks": [{"price": "0.46", "size": "11"}]
			},
			{
				"market": "condition-1",
				"asset_id": "down-token",
				"bids": [{"price": "0.54", "size": "10"}],
				"asks": [{"price": "0.56", "size": "11"}]
			}
		]`))
	}))
	defer server.Close()

	reader := NewReader(server.URL)
	batch, ok := reader.(BatchReader)
	if !ok {
		t.Fatalf("NewReader does not implement BatchReader")
	}
	books, err := batch.OrderBooks(context.Background(), []string{"up-token", "down-token"})
	if err != nil {
		t.Fatal(err)
	}

	if requests != 1 {
		t.Fatalf("requests=%d, want one batch request", requests)
	}
	if len(books) != 2 {
		t.Fatalf("books=%d", len(books))
	}
	if books[0].TokenID != "up-token" || books[0].Bids[0].Price != 0.44 || books[0].Asks[0].Price != 0.46 {
		t.Fatalf("up book=%#v", books[0])
	}
	if books[1].TokenID != "down-token" || books[1].Bids[0].Price != 0.54 || books[1].Asks[0].Price != 0.56 {
		t.Fatalf("down book=%#v", books[1])
	}
}
