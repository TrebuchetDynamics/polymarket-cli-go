package bridge_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/TrebuchetDynamics/polygolem/pkg/bridge"
)

// Example_supportedAssets demonstrates listing the Bridge's supported
// deposit assets. A test HTTP server stands in for the production Bridge
// so the example is hermetic and runs under `go test`.
func Example_supportedAssets() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"supportedAssets": [{
				"chainId": "137",
				"chainName": "Polygon",
				"token": {"name": "USDC", "symbol": "USDC", "address": "0x", "decimals": 6},
				"minCheckoutUsd": 5
			}]
		}`))
	}))
	defer server.Close()

	client := bridge.NewClient(server.URL, nil)
	resp, err := client.GetSupportedAssets(context.Background())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	first := resp.SupportedAssets[0]
	fmt.Printf("%s on %s, min=%.0f USD\n", first.Token.Symbol, first.ChainName, first.MinCheckoutUsd)
	// Output: USDC on Polygon, min=5 USD
}
