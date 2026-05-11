package plugins

import (
	"context"
	"errors"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

type noopMarketData struct{}

func (n *noopMarketData) Resolve(ctx context.Context, asset, timeframe string) (*types.Market, error) {
	return &types.Market{Slug: asset + "-" + timeframe}, nil
}

func (n *noopMarketData) Filter(ctx context.Context, market *types.Market) (bool, error) {
	return true, nil
}

type blockingRisk struct{}

func (b *blockingRisk) CheckOrder(ctx context.Context, order Order) error {
	return errors.New("blocked by plugin")
}

func TestMarketDataPlugin(t *testing.T) {
	var p MarketDataPlugin = &noopMarketData{}
	m, err := p.Resolve(context.Background(), "BTC", "5m")
	if err != nil {
		t.Fatal(err)
	}
	if m.Slug != "BTC-5m" {
		t.Fatalf("unexpected slug: %s", m.Slug)
	}
}

func TestRiskPlugin(t *testing.T) {
	var p RiskPlugin = &blockingRisk{}
	err := p.CheckOrder(context.Background(), Order{TokenID: "123", Side: "BUY"})
	if err == nil {
		t.Fatal("expected blocking error")
	}
}
