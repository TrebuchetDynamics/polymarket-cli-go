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
	"github.com/TrebuchetDynamics/polygolem/pkg/contracts"
	"github.com/TrebuchetDynamics/polygolem/pkg/data"
	"github.com/TrebuchetDynamics/polygolem/pkg/gamma"
	"github.com/TrebuchetDynamics/polygolem/pkg/marketdata"
	"github.com/TrebuchetDynamics/polygolem/pkg/orderbook"
	"github.com/TrebuchetDynamics/polygolem/pkg/orderresults"
	"github.com/TrebuchetDynamics/polygolem/pkg/relayer"
	"github.com/TrebuchetDynamics/polygolem/pkg/settlement"
	sdkstream "github.com/TrebuchetDynamics/polygolem/pkg/stream"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
	"github.com/TrebuchetDynamics/polygolem/pkg/universal"
	"github.com/TrebuchetDynamics/polygolem/pkg/wallet"
)

func TestPublicSDKSignatures(t *testing.T) {
	var clobClient *sdkclob.Client = sdkclob.NewClient(sdkclob.Config{})
	var clobConfig sdkclob.Config = sdkclob.Config{BuilderCode: "0x1111111111111111111111111111111111111111111111111111111111111111"}
	var clobMarkets func(*sdkclob.Client, context.Context, string) (*types.CLOBPaginatedMarkets, error) = (*sdkclob.Client).Markets
	var clobMarket func(*sdkclob.Client, context.Context, string) (*types.CLOBMarket, error) = (*sdkclob.Client).Market
	var clobMarketByToken func(*sdkclob.Client, context.Context, string) (*types.CLOBMarketByTokenResponse, error) = (*sdkclob.Client).MarketByToken
	var clobOrderBook func(*sdkclob.Client, context.Context, string) (*types.CLOBOrderBook, error) = (*sdkclob.Client).OrderBook
	var clobOrderBooks func(*sdkclob.Client, context.Context, []types.CLOBBookParams) ([]types.CLOBOrderBook, error) = (*sdkclob.Client).OrderBooks
	var clobTickSize func(*sdkclob.Client, context.Context, string) (*types.CLOBTickSize, error) = (*sdkclob.Client).TickSize
	var clobPriceHistory func(*sdkclob.Client, context.Context, *types.CLOBPriceHistoryParams) (*types.CLOBPriceHistory, error) = (*sdkclob.Client).PricesHistory
	var clobAPIKey sdkclob.APIKey
	var clobDeriveAPIKey func(*sdkclob.Client, context.Context, string) (sdkclob.APIKey, error) = (*sdkclob.Client).DeriveAPIKey
	var clobBalanceParams sdkclob.BalanceAllowanceParams
	var clobBalance func(*sdkclob.Client, context.Context, string, sdkclob.BalanceAllowanceParams) (*sdkclob.BalanceAllowanceResponse, error) = (*sdkclob.Client).BalanceAllowance
	var clobOrders func(*sdkclob.Client, context.Context, string) ([]sdkclob.OrderRecord, error) = (*sdkclob.Client).ListOrders
	var clobOrder func(*sdkclob.Client, context.Context, string, string) (*sdkclob.OrderRecord, error) = (*sdkclob.Client).Order
	var clobTrades func(*sdkclob.Client, context.Context, string) ([]sdkclob.TradeRecord, error) = (*sdkclob.Client).ListTrades
	var clobCancel func(*sdkclob.Client, context.Context, string, string) (*sdkclob.CancelOrdersResponse, error) = (*sdkclob.Client).CancelOrder
	var clobCancelMarketParams sdkclob.CancelMarketParams
	var clobCancelMarket func(*sdkclob.Client, context.Context, string, sdkclob.CancelMarketParams) (*sdkclob.CancelOrdersResponse, error) = (*sdkclob.Client).CancelMarket
	var clobCreateParams sdkclob.CreateOrderParams
	var clobCreate func(*sdkclob.Client, context.Context, string, sdkclob.CreateOrderParams) (*sdkclob.OrderPlacementResponse, error) = (*sdkclob.Client).CreateLimitOrder
	var clobMarketOrderParams sdkclob.MarketOrderParams
	var clobMarketOrder func(*sdkclob.Client, context.Context, string, sdkclob.MarketOrderParams) (*sdkclob.OrderPlacementResponse, error) = (*sdkclob.Client).CreateMarketOrder
	var streamClient *sdkstream.MarketClient = sdkstream.NewMarketClient(sdkstream.Config{})
	var streamConfig sdkstream.Config = sdkstream.DefaultConfig("")
	var streamConnect func(*sdkstream.MarketClient, context.Context) error = (*sdkstream.MarketClient).Connect
	var streamSubscribe func(*sdkstream.MarketClient, context.Context, []string) error = (*sdkstream.MarketClient).SubscribeAssets
	var streamClose func(*sdkstream.MarketClient) = (*sdkstream.MarketClient).Close
	var streamConnected func(*sdkstream.MarketClient) bool = (*sdkstream.MarketClient).IsConnected
	var streamBook sdkstream.BookMessage
	var streamPriceChange sdkstream.PriceChangeMessage
	var streamLastTrade sdkstream.LastTradeMessage
	var streamTickSize sdkstream.TickSizeChangeMessage
	var streamBestBidAsk sdkstream.BestBidAskMessage
	var streamNewMarket sdkstream.NewMarketMessage
	var streamMarketResolved sdkstream.MarketResolvedMessage
	var streamDeduplicator *sdkstream.Deduplicator = sdkstream.NewDeduplicator(100, 0)
	var marketDataTracker *marketdata.Tracker = marketdata.NewTracker()
	var marketDataSnapshot marketdata.Snapshot
	var marketDataBestBidAsk func(*marketdata.Tracker, sdkstream.BestBidAskMessage) marketdata.Snapshot = (*marketdata.Tracker).ApplyBestBidAsk
	var marketDataTickSize func(*marketdata.Tracker, sdkstream.TickSizeChangeMessage) marketdata.Snapshot = (*marketdata.Tracker).ApplyTickSizeChange
	var orderbookReader orderbook.Reader = orderbook.NewReader("")
	var orderbookSnapshot orderbook.OrderBook
	var orderbookLevel orderbook.Level
	var orderResultsSource orderresults.Source
	var orderResultsReport *orderresults.Report
	var orderResultsOptions orderresults.Options
	var orderResultsBuild func(context.Context, orderresults.DataReader, string, orderresults.Options) (*orderresults.Report, error) = orderresults.BuildReport
	var contractsRegistry contracts.Registry = contracts.PolygonMainnet()
	var contractStatus contracts.DeploymentStatus
	var contractDeployed func(context.Context, string, string) (contracts.DeploymentStatus, error) = contracts.ContractDeployed
	var depositWalletDeployed func(context.Context, string, string) (contracts.DeploymentStatus, error) = contracts.DepositWalletDeployed
	var redeemAdapterFor func(bool) string = contracts.RedeemAdapterFor
	var settlementPosition settlement.RedeemablePosition
	var settlementResult *settlement.RedeemResult
	var settlementReadiness *settlement.Readiness
	var settlementReadinessOptions settlement.ReadinessOptions
	var settlementAdapterApproval settlement.AdapterApproval
	var settlementFind func(context.Context, *data.Client, string) ([]settlement.RedeemablePosition, error) = settlement.FindRedeemable
	var settlementBuild func(settlement.RedeemablePosition) (relayer.DepositWalletCall, error) = settlement.BuildRedeemCall
	var settlementSubmit func(context.Context, *relayer.Client, string, []settlement.RedeemablePosition, int) (*settlement.RedeemResult, error) = settlement.SubmitRedeem
	var settlementCheck func(context.Context, *data.Client, string, string, settlement.ReadinessOptions) (*settlement.Readiness, error) = settlement.CheckReadiness
	var relayerClient *relayer.Client
	var relayerV2Key relayer.V2APIKey
	var relayerOnboardOptions relayer.OnboardOptions
	var relayerOnboard func(context.Context, *relayer.Client, string, relayer.OnboardOptions) (*relayer.OnboardResult, error) = relayer.OnboardDepositWallet
	var relayerNewV2 func(string, relayer.V2APIKey, int64) (*relayer.Client, error) = relayer.NewV2
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
	var universalConfig universal.Config = universal.Config{BuilderCode: "0x1111111111111111111111111111111111111111111111111111111111111111"}
	var universalCLOBMarkets func(*universal.Client, context.Context, string) (*types.CLOBPaginatedMarkets, error) = (*universal.Client).CLOBMarkets
	var universalCLOBMarket func(*universal.Client, context.Context, string) (*types.CLOBMarket, error) = (*universal.Client).CLOBMarket
	var universalCLOBMarketByToken func(*universal.Client, context.Context, string) (*types.CLOBMarketByTokenResponse, error) = (*universal.Client).CLOBMarketByToken
	var universalOrderBook func(*universal.Client, context.Context, string) (*types.CLOBOrderBook, error) = (*universal.Client).OrderBook
	var universalOrderBooks func(*universal.Client, context.Context, []types.CLOBBookParams) ([]types.CLOBOrderBook, error) = (*universal.Client).OrderBooks
	var universalTickSize func(*universal.Client, context.Context, string) (*types.CLOBTickSize, error) = (*universal.Client).TickSize
	var universalPriceHistory func(*universal.Client, context.Context, *types.CLOBPriceHistoryParams) (*types.CLOBPriceHistory, error) = (*universal.Client).PricesHistory
	var universalDeriveAPIKey func(*universal.Client, context.Context, string) (sdkclob.APIKey, error) = (*universal.Client).DeriveAPIKey
	var universalBalance func(*universal.Client, context.Context, string, sdkclob.BalanceAllowanceParams) (*sdkclob.BalanceAllowanceResponse, error) = (*universal.Client).BalanceAllowance
	var universalOrders func(*universal.Client, context.Context, string) ([]sdkclob.OrderRecord, error) = (*universal.Client).ListOrders
	var universalOrder func(*universal.Client, context.Context, string, string) (*sdkclob.OrderRecord, error) = (*universal.Client).Order
	var universalTrades func(*universal.Client, context.Context, string) ([]sdkclob.TradeRecord, error) = (*universal.Client).ListTrades
	var universalCancel func(*universal.Client, context.Context, string, string) (*sdkclob.CancelOrdersResponse, error) = (*universal.Client).CancelOrder
	var universalCancelMarket func(*universal.Client, context.Context, string, sdkclob.CancelMarketParams) (*sdkclob.CancelOrdersResponse, error) = (*universal.Client).CancelMarket
	var universalCreate func(*universal.Client, context.Context, string, sdkclob.CreateOrderParams) (*sdkclob.OrderPlacementResponse, error) = (*universal.Client).CreateLimitOrder
	var universalMarketOrder func(*universal.Client, context.Context, string, sdkclob.MarketOrderParams) (*sdkclob.OrderPlacementResponse, error) = (*universal.Client).CreateMarketOrder
	var universalStream func(*universal.Client) *sdkstream.MarketClient = (*universal.Client).StreamClient
	var universalStreamWithConfig func(*universal.Client, sdkstream.Config) *sdkstream.MarketClient = (*universal.Client).StreamClientWithConfig

	_, _, _, _, _, _, _, _, _ = clobClient, clobConfig, clobMarkets, clobMarket, clobMarketByToken, clobOrderBook, clobOrderBooks, clobTickSize, clobPriceHistory
	_, _, _, _, _, _, _, _, _, _ = clobAPIKey, clobDeriveAPIKey, clobBalanceParams, clobBalance, clobOrders, clobOrder, clobTrades, clobCancel, clobCancelMarketParams, clobCancelMarket
	_, _, _, _ = clobCreateParams, clobCreate, clobMarketOrderParams, clobMarketOrder
	_, _, _, _, _, _, _, _, _, _ = streamClient, streamConfig, streamConnect, streamSubscribe, streamClose, streamConnected, streamBook, streamPriceChange, streamLastTrade, streamTickSize
	_, _, _, _, _, _ = streamBestBidAsk, streamNewMarket, streamMarketResolved, streamDeduplicator, marketDataTracker, marketDataSnapshot
	_, _ = marketDataBestBidAsk, marketDataTickSize
	_, _, _, _, _, _, _ = orderbookReader, orderbookSnapshot, orderbookLevel, orderResultsSource, orderResultsReport, orderResultsOptions, orderResultsBuild
	_, _, _, _, _ = contractsRegistry, contractStatus, contractDeployed, depositWalletDeployed, redeemAdapterFor
	_, _, _, _, _, _, _, _, _ = settlementPosition, settlementResult, settlementReadiness, settlementReadinessOptions, settlementAdapterApproval, settlementFind, settlementBuild, settlementSubmit, settlementCheck
	_, _, _, _, _ = relayerClient, relayerV2Key, relayerOnboardOptions, relayerOnboard, relayerNewV2
	_, _, _, _ = dataPositions, universalPositions, dataLeaderboard, universalLiveVolume
	_, _, _, _, _, _, _ = gammaMarkets, gammaSearch, gammaComments, universalMarkets, universalSearch, universalComments, universalConfig
	_, _, _, _, _, _, _ = universalCLOBMarkets, universalCLOBMarket, universalCLOBMarketByToken, universalOrderBook, universalOrderBooks, universalTickSize, universalPriceHistory
	_, _, _, _, _, _, _, _, _ = universalDeriveAPIKey, universalBalance, universalOrders, universalOrder, universalTrades, universalCancel, universalCancelMarket, universalCreate, universalMarketOrder
	_, _ = universalStream, universalStreamWithConfig

	var walletProxy func(string) string = wallet.DeriveProxyWallet
	var walletSafe func(string) string = wallet.DeriveSafeWallet
	var walletReady func(int64, string) wallet.ReadyInfo = wallet.Readiness
	_ = walletProxy
	_ = walletSafe
	_ = walletReady
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
