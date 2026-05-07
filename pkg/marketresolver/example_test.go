package marketresolver_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/TrebuchetDynamics/polygolem/pkg/marketresolver"
)

// Example_resolveTokenIDsAt demonstrates resolving the up/down token IDs
// for a deterministic Polymarket crypto window slug. A test HTTP server
// stands in for the production Gamma API so the example is hermetic and
// runs under `go test`.
func Example_resolveTokenIDsAt() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events" && r.URL.Query().Get("slug") == "btc-updown-5m-1778114700" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{
				"id": "event-1",
				"slug": "btc-updown-5m-1778114700",
				"active": true,
				"closed": false,
				"markets": [{
					"id": "market-1",
					"conditionId": "condition-1",
					"slug": "btc-updown-5m-1778114700",
					"question": "Bitcoin Up or Down",
					"outcomes": ["Up", "Down"],
					"active": true,
					"closed": false,
					"enableOrderBook": true,
					"acceptingOrders": true,
					"clobTokenIds": "[\"up-token\", \"down-token\"]"
				}]
			}]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	resolver := marketresolver.NewResolver(server.URL)
	got := resolver.ResolveTokenIDsAt(context.Background(), "BTC", "5m", time.Unix(1778114700, 0).UTC())
	fmt.Printf("status=%s up=%s down=%s\n", got.Status, got.UpTokenID, got.DownTokenID)
	// Output: status=available up=up-token down=down-token
}
