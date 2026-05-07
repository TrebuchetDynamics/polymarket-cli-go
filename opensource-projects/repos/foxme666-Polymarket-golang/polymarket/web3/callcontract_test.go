package web3

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// TestCallContractPlainStrategy 验证 plain 策略（无 gas 字段）在正常节点上工作
func TestCallContractPlainStrategy(t *testing.T) {
	var capturedBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000001234567890abcdef1234567890abcdef12345678"}`))
	}))
	defer ts.Close()

	rpcClient, _ := rpc.Dial(ts.URL)
	defer rpcClient.Close()

	base := &BaseWeb3Client{
		client:    ethclient.NewClient(rpcClient),
		rpcClient: rpcClient,
		chainID:   137,
	}

	to := common.HexToAddress("0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E")
	result, err := base.callContract(context.Background(), &to, common.FromHex("0x06fdde03"))
	if err != nil {
		t.Fatalf("callContract failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	// 验证使用了 plain 策略（缓存生效）
	if base.ethCallMode != ethCallPlain {
		t.Errorf("expected ethCallPlain strategy, got %d", base.ethCallMode)
	}

	// 验证 JSON 参数不含 gas 字段
	var req struct{ Params []json.RawMessage }
	json.Unmarshal(capturedBody, &req)
	var args map[string]interface{}
	json.Unmarshal(req.Params[0], &args)

	for _, f := range []string{"gasPrice", "maxFeePerGas", "maxPriorityFeePerGas", "from"} {
		if _, ok := args[f]; ok {
			t.Errorf("plain strategy should NOT send '%s'", f)
		}
	}
}

// TestCallContractAutoDetectFallsThrough 验证当 plain 策略失败时自动尝试其他策略
func TestCallContractAutoDetectFallsThrough(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")

		// 检查是否包含 maxFeePerGas（第二次尝试 = EIP-1559 策略）
		if strings.Contains(string(body), "maxFeePerGas") {
			// EIP-1559 策略成功
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000001234567890abcdef1234567890abcdef12345678"}`))
			return
		}

		// plain 策略失败（模拟 baseFee 错误）
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"max fee per gas less than block base fee"}}`))
	}))
	defer ts.Close()

	rpcClient, _ := rpc.Dial(ts.URL)
	defer rpcClient.Close()

	base := &BaseWeb3Client{
		client:    ethclient.NewClient(rpcClient),
		rpcClient: rpcClient,
		chainID:   137,
	}

	result, err := base.callContract(context.Background(),
		&common.Address{}, common.FromHex("0x06fdde03"))
	if err != nil {
		t.Fatalf("callContract should succeed with EIP-1559 fallback: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	// 应该缓存 EIP-1559 策略
	if base.ethCallMode != ethCallEIP1559 {
		t.Errorf("expected ethCallEIP1559 strategy, got %d", base.ethCallMode)
	}

	// 验证缓存生效：第二次调用直接使用 EIP-1559
	callCount = 0
	_, err = base.callContract(context.Background(),
		&common.Address{}, common.FromHex("0x06fdde03"))
	if err != nil {
		t.Fatalf("cached call failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("cached call should only make 1 RPC request, got %d", callCount)
	}
}

// TestCallContractFallbackRPC 验证主 RPC 失败时降级到备用 RPC
func TestCallContractFallbackRPC(t *testing.T) {
	// 主 RPC：所有请求都失败
	primaryTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"max fee per gas less than block base fee"}}`))
	}))
	defer primaryTS.Close()

	// 备用 RPC：成功
	fallbackTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000001234567890abcdef1234567890abcdef12345678"}`))
	}))
	defer fallbackTS.Close()

	// 临时覆盖 fallback RPCs
	origFallbacks := polygonFallbackRPCs[999]
	polygonFallbackRPCs[999] = []string{fallbackTS.URL}
	defer func() {
		if origFallbacks == nil {
			delete(polygonFallbackRPCs, 999)
		} else {
			polygonFallbackRPCs[999] = origFallbacks
		}
	}()

	rpcClient, _ := rpc.Dial(primaryTS.URL)
	defer rpcClient.Close()

	base := &BaseWeb3Client{
		client:    ethclient.NewClient(rpcClient),
		rpcClient: rpcClient,
		chainID:   999,
	}

	result, err := base.callContract(context.Background(),
		&common.Address{}, common.FromHex("0x06fdde03"))
	if err != nil {
		t.Fatalf("callContract should succeed with fallback RPC: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	// 验证使用了 fallback 客户端
	if base.ethCallClient == base.rpcClient {
		t.Error("expected ethCallClient to be fallback, not primary")
	}
	if base.fallbackRPC == nil {
		t.Error("expected fallbackRPC to be set")
	}
}

// TestCallContractError 验证所有策略和所有 RPC 都失败时返回错误
func TestCallContractError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"execution reverted"}}`))
	}))
	defer ts.Close()

	rpcClient, _ := rpc.Dial(ts.URL)
	defer rpcClient.Close()

	base := &BaseWeb3Client{
		client:    ethclient.NewClient(rpcClient),
		rpcClient: rpcClient,
		chainID:   12345, // 无效的 chainID，没有 fallback RPC
	}

	_, err := base.callContract(context.Background(),
		&common.Address{}, []byte{0x01})
	if err == nil {
		t.Fatal("expected error when all strategies fail")
	}
}
