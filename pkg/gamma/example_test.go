package gamma_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/TrebuchetDynamics/polygolem/pkg/gamma"
)

// Example_healthCheck demonstrates a Gamma readiness probe. A test HTTP
// server stands in for the production Gamma API so the example is
// hermetic and runs under `go test`.
func Example_healthCheck() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data": "ok"}`))
	}))
	defer server.Close()

	client := gamma.NewClient(server.URL)
	resp, err := client.HealthCheck(context.Background())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("data:", resp.Data)
	// Output: data: ok
}
