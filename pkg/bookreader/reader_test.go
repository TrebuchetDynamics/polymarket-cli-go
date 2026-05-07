package bookreader

import (
	"context"
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
