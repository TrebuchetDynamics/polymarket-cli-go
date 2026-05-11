package enabletrading

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	sdkclob "github.com/TrebuchetDynamics/polygolem/pkg/clob"
	sdkrelayer "github.com/TrebuchetDynamics/polygolem/pkg/relayer"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

const (
	testPrivateKey    = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	testEOA           = "0x2c7536E3605D9C16a7a3D7b1898e529396a65c23"
	testDepositWallet = "0x21999a074344610057c9b2B362332388a44502D4"
)

func TestBuildClobAuthTypedDataMatchesPolymarketUIShape(t *testing.T) {
	td, err := BuildClobAuthTypedData(ClobAuthParams{
		Address:   strings.ToLower(testEOA),
		ChainID:   137,
		Timestamp: "1778372101",
		Nonce:     0,
	})
	if err != nil {
		t.Fatalf("BuildClobAuthTypedData: %v", err)
	}

	if td.PrimaryType != "ClobAuth" {
		t.Fatalf("primaryType=%q", td.PrimaryType)
	}
	if td.Domain.Name != "ClobAuthDomain" || td.Domain.Version != "1" || td.Domain.ChainID != 137 {
		t.Fatalf("unexpected domain: %+v", td.Domain)
	}
	if td.Message.Address != strings.ToLower(testEOA) {
		t.Fatalf("address=%q want lowercase EOA", td.Message.Address)
	}
	if td.Message.Timestamp != "1778372101" || td.Message.Nonce != 0 {
		t.Fatalf("unexpected message timing: %+v", td.Message)
	}
	if td.Message.Message != ClobAuthControlMessage {
		t.Fatalf("message=%q", td.Message.Message)
	}
	assertTypeFields(t, td.Types["EIP712Domain"], []string{"name:string", "version:string", "chainId:uint256"})
	assertTypeFields(t, td.Types["ClobAuth"], []string{"address:address", "timestamp:string", "nonce:uint256", "message:string"})

	raw, err := json.Marshal(td)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Domain struct {
			ChainID int64 `json:"chainId"`
		} `json:"domain"`
		Message struct {
			Nonce uint64 `json:"nonce"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Domain.ChainID != 137 || decoded.Message.Nonce != 0 {
		t.Fatalf("json numeric fields changed: %s", raw)
	}
}

func TestSignClobAuthTypedDataRecoversEOA(t *testing.T) {
	td, err := BuildClobAuthTypedData(ClobAuthParams{
		Address:   testEOA,
		ChainID:   137,
		Timestamp: "1778372101",
		Nonce:     0,
	})
	if err != nil {
		t.Fatal(err)
	}

	sig, err := SignClobAuthTypedData(testPrivateKey, td)
	if err != nil {
		t.Fatalf("SignClobAuthTypedData: %v", err)
	}
	if !strings.HasPrefix(sig, "0x") || len(sig) != 132 {
		t.Fatalf("signature shape=%q", sig)
	}

	hash, err := HashClobAuthTypedData(td)
	if err != nil {
		t.Fatal(err)
	}
	recovered := recoverAddress(t, hash, sig)
	if !strings.EqualFold(recovered, testEOA) {
		t.Fatalf("recovered=%s want %s", recovered, testEOA)
	}
}

func TestBuildEnableTradingApprovalCallsMatchObservedUIBatch(t *testing.T) {
	calls := BuildEnableTradingApprovalCalls()
	if len(calls) != 2 {
		t.Fatalf("len=%d want 2", len(calls))
	}

	assertApproveCall(t, calls[0],
		"0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB",
		"0x4D97DCd97eC945f40cF65F87097ACe5EA0476045",
	)
	assertApproveCall(t, calls[1],
		"0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
		"0x93070a847efEf7F70739046A929D47a521F5B8ee",
	)
}

func TestBuildEnableTradingApprovalBatchTypedDataMatchesObservedShape(t *testing.T) {
	calls := BuildEnableTradingApprovalCalls()
	td, err := BuildEnableTradingApprovalBatchTypedData(ApprovalBatchParams{
		DepositWallet: testDepositWallet,
		ChainID:       137,
		Nonce:         "6",
		Deadline:      "1778373936",
		Calls:         calls,
	})
	if err != nil {
		t.Fatalf("BuildEnableTradingApprovalBatchTypedData: %v", err)
	}

	if td.PrimaryType != "Batch" {
		t.Fatalf("primaryType=%q", td.PrimaryType)
	}
	if td.Domain.Name != "DepositWallet" || td.Domain.Version != "1" || td.Domain.ChainID != 137 {
		t.Fatalf("unexpected domain: %+v", td.Domain)
	}
	if !strings.EqualFold(td.Domain.VerifyingContract, testDepositWallet) {
		t.Fatalf("verifyingContract=%s want %s", td.Domain.VerifyingContract, testDepositWallet)
	}
	if !strings.EqualFold(td.Message.Wallet, testDepositWallet) || td.Message.Nonce != "6" || td.Message.Deadline != "1778373936" {
		t.Fatalf("unexpected message: %+v", td.Message)
	}
	if len(td.Message.Calls) != 2 {
		t.Fatalf("calls=%d", len(td.Message.Calls))
	}
	assertTypeFields(t, td.Types["EIP712Domain"], []string{"name:string", "version:string", "chainId:uint256", "verifyingContract:address"})
	assertTypeFields(t, td.Types["Call"], []string{"target:address", "value:uint256", "data:bytes"})
	assertTypeFields(t, td.Types["Batch"], []string{"wallet:address", "nonce:uint256", "deadline:uint256", "calls:Call[]"})
}

func TestEnableTradingHeadlessDryRunDoesNotSignOrSubmit(t *testing.T) {
	result, err := EnableTradingHeadless(context.Background(), EnableTradingParams{
		OwnerPrivateKey:       testPrivateKey,
		DepositWalletAddress:  testDepositWallet,
		CreateOrDeriveCLOBKey: true,
		ApproveTokens:         true,
		MaxApproval:           true,
		DryRun:                true,
		ClobAuthTimestamp:     "1778372101",
		WalletNonce:           "6",
		ApprovalDeadline:      "1778373936",
	})
	if err != nil {
		t.Fatalf("EnableTradingHeadless dry-run: %v", err)
	}

	if result.CLOBAuthSigned || result.APIKeysCreatedOrDerived || result.TokenApprovalsSigned || result.TokenApprovalsSubmitted || result.ReadyToTrade {
		t.Fatalf("dry-run performed live work: %+v", result)
	}
	if result.ClobAuthTypedData == nil || result.ApprovalBatchTypedData == nil {
		t.Fatalf("dry-run should return typed data: %+v", result)
	}
	if len(result.PlannedActions) == 0 {
		t.Fatalf("planned actions missing: %+v", result)
	}
}

func TestEnableTradingHeadlessLivePathPollsDeployAndSubmitsApprovals(t *testing.T) {
	relayerClient := &fakeWalletRelayer{nonce: "11"}
	clobClient := &fakeCLOBKeyClient{key: sdkclob.APIKey{Key: "api-key", Secret: "api-secret", Passphrase: "api-pass"}}

	result, err := EnableTradingHeadless(context.Background(), EnableTradingParams{
		OwnerPrivateKey:       testPrivateKey,
		DepositWalletAddress:  testDepositWallet,
		DeployIfNeeded:        true,
		CreateOrDeriveCLOBKey: true,
		ApproveTokens:         true,
		MaxApproval:           true,
		ClobAuthTimestamp:     "1778372101",
		ApprovalDeadline:      sdkrelayer.BuildDeadline(240),
		CLOB:                  clobClient,
		Relayer:               relayerClient,
	})
	if err != nil {
		t.Fatalf("EnableTradingHeadless live path: %v", err)
	}

	if !relayerClient.submittedCreate || !relayerClient.polledCreate {
		t.Fatalf("deploy was not submitted and polled: %+v", relayerClient)
	}
	if !clobClient.called {
		t.Fatal("CLOB key client was not called")
	}
	if !relayerClient.submittedBatch {
		t.Fatal("approval batch was not submitted")
	}
	if !result.Deployed || !result.CLOBAuthSigned || !result.APIKeysCreatedOrDerived || !result.TokenApprovalsSigned || !result.TokenApprovalsSubmitted || !result.ReadyToTrade {
		t.Fatalf("unexpected readiness result: %+v", result)
	}
	if len(result.TxHashes) != 2 || result.TxHashes[0] != "0xdeploy" || result.TxHashes[1] != "0xbatch" {
		t.Fatalf("tx hashes=%v", result.TxHashes)
	}
}

func TestSafetyValidationFailsClosed(t *testing.T) {
	if _, err := BuildClobAuthTypedData(ClobAuthParams{
		Address:   testEOA,
		ChainID:   1,
		Timestamp: "1778372101",
		Nonce:     0,
	}); err == nil {
		t.Fatal("expected wrong chain to fail")
	}

	if _, err := BuildEnableTradingApprovalBatchTypedData(ApprovalBatchParams{
		DepositWallet: testDepositWallet,
		ChainID:       137,
		Nonce:         "6",
		Deadline:      "1778373936",
		Calls: []DepositWalletCall{{
			Target: "0x0000000000000000000000000000000000000001",
			Value:  "0",
			Data:   "0x095ea7b30000000000000000000000004d97dcd97ec945f40cf65f87097ace5ea0476045ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		}},
	}); err == nil {
		t.Fatal("expected unknown approval target to fail")
	}
}

func assertTypeFields(t *testing.T, got []TypedDataField, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("field len=%d want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		pair := got[i].Name + ":" + got[i].Type
		if pair != want[i] {
			t.Fatalf("field %d=%s want %s", i, pair, want[i])
		}
	}
}

func assertApproveCall(t *testing.T, call DepositWalletCall, wantTarget, wantSpender string) {
	t.Helper()
	if !strings.EqualFold(call.Target, wantTarget) {
		t.Fatalf("target=%s want %s", call.Target, wantTarget)
	}
	data := strings.ToLower(strings.TrimPrefix(call.Data, "0x"))
	if !strings.HasPrefix(data, "095ea7b3") {
		t.Fatalf("data selector=%s", call.Data)
	}
	if !strings.Contains(data, strings.ToLower(strings.TrimPrefix(wantSpender, "0x"))) {
		t.Fatalf("spender %s missing from calldata %s", wantSpender, call.Data)
	}
	if !strings.HasSuffix(data, strings.Repeat("f", 64)) {
		t.Fatalf("max uint amount missing from calldata %s", call.Data)
	}
}

func recoverAddress(t *testing.T, hash []byte, sigHex string) string {
	t.Helper()
	sig := common.FromHex(sigHex)
	if len(sig) != 65 {
		t.Fatalf("sig len=%d", len(sig))
	}
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	pub, err := ethcrypto.SigToPub(hash, sig)
	if err != nil {
		t.Fatal(err)
	}
	return ethcrypto.PubkeyToAddress(*pub).Hex()
}

type fakeWalletRelayer struct {
	deployed        bool
	nonce           string
	submittedCreate bool
	polledCreate    bool
	submittedBatch  bool
}

func (f *fakeWalletRelayer) IsDeployed(ctx context.Context, ownerAddress string) (bool, error) {
	return f.deployed, nil
}

func (f *fakeWalletRelayer) SubmitWalletCreate(ctx context.Context, ownerAddress string) (*sdkrelayer.RelayerTransaction, error) {
	f.submittedCreate = true
	return &sdkrelayer.RelayerTransaction{TransactionID: "deploy-1", State: string(sdkrelayer.StateNew), Type: "WALLET-CREATE"}, nil
}

func (f *fakeWalletRelayer) PollTransaction(ctx context.Context, txID string, maxAttempts int, interval time.Duration) (*sdkrelayer.RelayerTransaction, error) {
	f.polledCreate = true
	f.deployed = true
	return &sdkrelayer.RelayerTransaction{TransactionID: txID, TransactionHash: "0xdeploy", State: string(sdkrelayer.StateMined), Type: "WALLET-CREATE"}, nil
}

func (f *fakeWalletRelayer) GetNonce(ctx context.Context, ownerAddress string) (string, error) {
	return f.nonce, nil
}

func (f *fakeWalletRelayer) SubmitWalletBatch(ctx context.Context, ownerAddress, walletAddress, nonce, signature, deadline string, calls []DepositWalletCall) (*sdkrelayer.RelayerTransaction, error) {
	f.submittedBatch = true
	return &sdkrelayer.RelayerTransaction{TransactionID: "batch-1", TransactionHash: "0xbatch", State: string(sdkrelayer.StateMined), Type: "WALLET"}, nil
}

type fakeCLOBKeyClient struct {
	called bool
	key    sdkclob.APIKey
}

func (f *fakeCLOBKeyClient) CreateOrDeriveAPIKey(ctx context.Context, privateKey string) (sdkclob.APIKey, error) {
	f.called = true
	return f.key, nil
}
