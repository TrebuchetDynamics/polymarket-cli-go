package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPublicDataAPIDoesNotRequireInternalImports(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
	tempDir := t.TempDir()

	writeFile(t, filepath.Join(tempDir, "go.mod"), `module example.com/polygolem-public-consumer

go 1.25.0

require github.com/TrebuchetDynamics/polygolem v0.0.0

replace github.com/TrebuchetDynamics/polygolem => `+repoRoot+`
`)
	writeFile(t, filepath.Join(tempDir, "public_sdk_test.go"), `package publicconsumer

import (
	"context"
	"testing"

	sdkclob "github.com/TrebuchetDynamics/polygolem/pkg/clob"
	"github.com/TrebuchetDynamics/polygolem/pkg/data"
	"github.com/TrebuchetDynamics/polygolem/pkg/bookreader"
	"github.com/TrebuchetDynamics/polygolem/pkg/gamma"
	"github.com/TrebuchetDynamics/polygolem/pkg/orderbook"
	sdkstream "github.com/TrebuchetDynamics/polygolem/pkg/stream"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
	"github.com/TrebuchetDynamics/polygolem/pkg/universal"
)

func TestPublicSDKSignatures(t *testing.T) {
	var clobClient *sdkclob.Client = sdkclob.NewClient(sdkclob.Config{})
	var clobMarkets func(*sdkclob.Client, context.Context, string) (*types.CLOBPaginatedMarkets, error) = (*sdkclob.Client).Markets
	var clobMarket func(*sdkclob.Client, context.Context, string) (*types.CLOBMarket, error) = (*sdkclob.Client).Market
	var clobOrderBook func(*sdkclob.Client, context.Context, string) (*types.CLOBOrderBook, error) = (*sdkclob.Client).OrderBook
	var clobOrderBooks func(*sdkclob.Client, context.Context, []types.CLOBBookParams) ([]types.CLOBOrderBook, error) = (*sdkclob.Client).OrderBooks
	var clobTickSize func(*sdkclob.Client, context.Context, string) (*types.CLOBTickSize, error) = (*sdkclob.Client).TickSize
	var clobPriceHistory func(*sdkclob.Client, context.Context, *types.CLOBPriceHistoryParams) (*types.CLOBPriceHistory, error) = (*sdkclob.Client).PricesHistory
	var streamClient *sdkstream.MarketClient = sdkstream.NewMarketClient(sdkstream.Config{})
	var streamConfig sdkstream.Config = sdkstream.DefaultConfig("")
	var streamConnect func(*sdkstream.MarketClient, context.Context) error = (*sdkstream.MarketClient).Connect
	var streamSubscribe func(*sdkstream.MarketClient, context.Context, []string) error = (*sdkstream.MarketClient).SubscribeAssets
	var streamClose func(*sdkstream.MarketClient) = (*sdkstream.MarketClient).Close
	var streamConnected func(*sdkstream.MarketClient) bool = (*sdkstream.MarketClient).IsConnected
	var streamBook sdkstream.BookMessage
	var streamPriceChange sdkstream.PriceChangeMessage
	var streamLastTrade sdkstream.LastTradeMessage
	var streamDeduplicator *sdkstream.Deduplicator = sdkstream.NewDeduplicator(100, 0)
	var orderbookReader orderbook.Reader = orderbook.NewReader("")
	var orderbookSnapshot orderbook.OrderBook
	var orderbookLevel orderbook.Level
	var legacyReader bookreader.Reader = bookreader.NewReader("")
	var dataPositions func(*data.Client, context.Context, string) ([]types.Position, error) = (*data.Client).CurrentPositions
	var universalPositions func(*universal.Client, context.Context, string) ([]types.Position, error) = (*universal.Client).CurrentPositions
	var dataLeaderboard func(*data.Client, context.Context, int) ([]types.LeaderboardRow, error) = (*data.Client).TraderLeaderboard
	var universalLiveVolume func(*universal.Client, context.Context, int) (*types.LiveVolumeResponse, error) = (*universal.Client).LiveVolume
	var gammaMarkets func(*gamma.Client, context.Context, *types.GetMarketsParams) ([]types.Market, error) = (*gamma.Client).Markets
	var gammaSearch func(*gamma.Client, context.Context, *types.SearchParams) (*types.SearchResponse, error) = (*gamma.Client).Search
	var gammaComments func(*gamma.Client, context.Context, *types.CommentQuery) ([]types.Comment, error) = (*gamma.Client).Comments
	var universalMarkets func(*universal.Client, context.Context, *types.GetMarketsParams) ([]types.Market, error) = (*universal.Client).Markets
	var universalSearch func(*universal.Client, context.Context, *types.SearchParams) (*types.SearchResponse, error) = (*universal.Client).Search
	var universalComments func(*universal.Client, context.Context, *types.CommentQuery) ([]types.Comment, error) = (*universal.Client).Comments
	var universalCLOBMarkets func(*universal.Client, context.Context, string) (*types.CLOBPaginatedMarkets, error) = (*universal.Client).CLOBMarkets
	var universalCLOBMarket func(*universal.Client, context.Context, string) (*types.CLOBMarket, error) = (*universal.Client).CLOBMarket
	var universalOrderBook func(*universal.Client, context.Context, string) (*types.CLOBOrderBook, error) = (*universal.Client).OrderBook
	var universalOrderBooks func(*universal.Client, context.Context, []types.CLOBBookParams) ([]types.CLOBOrderBook, error) = (*universal.Client).OrderBooks
	var universalTickSize func(*universal.Client, context.Context, string) (*types.CLOBTickSize, error) = (*universal.Client).TickSize
	var universalPriceHistory func(*universal.Client, context.Context, *types.CLOBPriceHistoryParams) (*types.CLOBPriceHistory, error) = (*universal.Client).PricesHistory
	var universalStream func(*universal.Client) *sdkstream.MarketClient = (*universal.Client).StreamClient
	var universalStreamWithConfig func(*universal.Client, sdkstream.Config) *sdkstream.MarketClient = (*universal.Client).StreamClientWithConfig

	_, _, _, _, _, _, _ = clobClient, clobMarkets, clobMarket, clobOrderBook, clobOrderBooks, clobTickSize, clobPriceHistory
	_, _, _, _, _, _, _, _, _, _ = streamClient, streamConfig, streamConnect, streamSubscribe, streamClose, streamConnected, streamBook, streamPriceChange, streamLastTrade, streamDeduplicator
	_, _, _, _ = orderbookReader, orderbookSnapshot, orderbookLevel, legacyReader
	_, _, _, _ = dataPositions, universalPositions, dataLeaderboard, universalLiveVolume
	_, _, _, _, _, _ = gammaMarkets, gammaSearch, gammaComments, universalMarkets, universalSearch, universalComments
	_, _, _, _, _, _ = universalCLOBMarkets, universalCLOBMarket, universalOrderBook, universalOrderBooks, universalTickSize, universalPriceHistory
	_, _ = universalStream, universalStreamWithConfig
}
`)

	cmd := exec.Command("go", "test", "-mod=mod", "./...")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("external consumer compile failed: %v\n%s", err, out)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
