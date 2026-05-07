package bookreader_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/TrebuchetDynamics/polygolem/pkg/bookreader"
)

// Example_orderBook demonstrates fetching a CLOB order-book snapshot for a
// token ID. A test HTTP server stands in for the production CLOB so the
// example is hermetic and runnable with `go test ./pkg/bookreader`.
func Example_orderBook() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"market": "condition-1",
			"asset_id": "token-1",
			"bids": [{"price": "0.42", "size": "10"}],
			"asks": [{"price": "0.58", "size": "10"}]
		}`))
	}))
	defer server.Close()

	reader := bookreader.NewReader(server.URL)
	book, err := reader.OrderBook(context.Background(), "token-1")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("best bid=%.2f best ask=%.2f\n", book.Bids[0].Price, book.Asks[0].Price)
	// Output: best bid=0.42 best ask=0.58
}
