package enabletrading

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/polygolem/internal/auth"
	"github.com/TrebuchetDynamics/polygolem/internal/relayer"
)

const testPrivateKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

func TestBuildClobAuthTypedDataShape(t *testing.T) {
	typed := BuildClobAuthTypedData("0x2c7536E3605D9C16a7a3D7b1898e529396a65c23", "1700000000", 0)

	if typed.PrimaryType != "ClobAuth" {
		t.Fatalf("primaryType=%s", typed.PrimaryType)
	}
	if typed.Domain.Name != "ClobAuthDomain" || typed.Domain.Version != "1" || typed.Domain.ChainID != PolygonChainID {
		t.Fatalf("domain=%+v", typed.Domain)
	}
	if typed.Message["address"] != "0x2c7536E3605D9C16a7a3D7b1898e529396a65c23" {
		t.Fatalf("address=%v", typed.Message["address"])
	}
	if typed.Message["timestamp"] != "1700000000" {
		t.Fatalf("timestamp=%v", typed.Message["timestamp"])
	}
	if typed.Message["nonce"] != "0" {
		t.Fatalf("nonce=%v", typed.Message["nonce"])
	}
	if typed.Message["message"] != ClobAuthAttestationMessage {
		t.Fatalf("message=%v", typed.Message["message"])
	}
	if len(typed.Types["ClobAuth"]) != 4 {
		t.Fatalf("ClobAuth type fields=%d", len(typed.Types["ClobAuth"]))
	}
}

func TestSignClobAuthTypedDataRejectsWrongChain(t *testing.T) {
	signer, err := auth.NewPrivateKeySigner(testPrivateKey, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = SignClobAuthTypedData(signer, signer.Address(), "1700000000", 0)
	if err == nil || !strings.Contains(err.Error(), "chainId 137") {
		t.Fatalf("err=%v", err)
	}
}

func TestBuildDepositWalletApprovalBatchTypedDataMatchesObservedShape(t *testing.T) {
	wallet := "0x1111111111111111111111111111111111111111"
	calls := ObservedEnableTradingApprovalCalls()
	typed, err := BuildDepositWalletApprovalBatchTypedData(wallet, "7", "1800000000", calls)
	if err != nil {
		t.Fatal(err)
	}
	if typed.PrimaryType != "Batch" {
		t.Fatalf("primaryType=%s", typed.PrimaryType)
	}
	if typed.Domain.Name != "DepositWallet" || typed.Domain.Version != "1" || typed.Domain.ChainID != PolygonChainID {
		t.Fatalf("domain=%+v", typed.Domain)
	}
	if !strings.EqualFold(typed.Domain.VerifyingContract, wallet) {
		t.Fatalf("verifyingContract=%s", typed.Domain.VerifyingContract)
	}
	if typed.Message["wallet"] != wallet || typed.Message["nonce"] != "7" || typed.Message["deadline"] != "1800000000" {
		t.Fatalf("message=%+v", typed.Message)
	}
	gotCalls := typed.Message["calls"].([]map[string]string)
	if len(gotCalls) != 2 {
		t.Fatalf("calls=%d", len(gotCalls))
	}
	assertERC20Approve(t, gotCalls[0], "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB", "0x4d97dcd97ec945f40cf65f87097ace5ea0476045")
	assertERC20Approve(t, gotCalls[1], "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174", "0x93070a847efef7f70739046a929d47a521f5b8ee")
}

func TestBuildDepositWalletApprovalBatchRejectsMismatchedWallet(t *testing.T) {
	_, err := BuildDepositWalletApprovalBatchTypedData("", "7", "1800000000", ObservedEnableTradingApprovalCalls())
	if err == nil {
		t.Fatal("expected error for missing wallet")
	}
}

func TestSignDepositWalletApprovalBatchRejectsExpiredDeadline(t *testing.T) {
	signer, err := auth.NewPrivateKeySigner(testPrivateKey, PolygonChainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = SignDepositWalletApprovalBatch(signer, signer.Address(), "7", "1", ObservedEnableTradingApprovalCalls())
	if err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("err=%v", err)
	}
}

func TestEnableTradingHeadlessDryRunDoesNotSignOrSubmit(t *testing.T) {
	signer, err := auth.NewPrivateKeySigner(testPrivateKey, PolygonChainID)
	if err != nil {
		t.Fatal(err)
	}
	wallet := "0x1111111111111111111111111111111111111111"
	relayerClient := &fakeRelayer{deployed: true, nonce: "9"}
	clobClient := &fakeClob{key: auth.APIKey{Key: "api-key", Secret: "api-secret", Passphrase: "api-pass"}}

	result, err := EnableTradingHeadless(context.Background(), Options{
		Signer:               signer,
		OwnerPrivateKey:      testPrivateKey,
		DepositWalletAddress: wallet,
		DeployIfNeeded:       true,
		CreateOrDeriveAPIKey: true,
		ApproveTokens:        true,
		MaxApproval:          true,
		DryRun:               true,
		Relayer:              relayerClient,
		CLOB:                 clobClient,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.DepositWalletAddress != wallet {
		t.Fatalf("wallet=%s", result.DepositWalletAddress)
	}
	if result.ClobAuthSigned || result.APIKeysCreated || result.TokenApprovalsSigned || result.TokenApprovalsSubmitted {
		t.Fatalf("dry-run mutated result: %+v", result)
	}
	if relayerClient.submittedBatch || clobClient.called {
		t.Fatalf("dry-run called side effects: relayer=%v clob=%v", relayerClient.submittedBatch, clobClient.called)
	}
	if result.PlannedClobAuth == nil || result.PlannedApprovalBatch == nil {
		t.Fatalf("dry-run missing planned typed data: %+v", result)
	}
}

func TestEnableTradingHeadlessIdempotentWhenReady(t *testing.T) {
	signer, err := auth.NewPrivateKeySigner(testPrivateKey, PolygonChainID)
	if err != nil {
		t.Fatal(err)
	}
	wallet := "0x1111111111111111111111111111111111111111"
	relayerClient := &fakeRelayer{deployed: true, approvalsReady: true, nonce: "9"}
	clobClient := &fakeClob{existing: true, key: auth.APIKey{Key: "api-key", Secret: "api-secret", Passphrase: "api-pass"}}

	result, err := EnableTradingHeadless(context.Background(), Options{
		Signer:               signer,
		OwnerPrivateKey:      testPrivateKey,
		DepositWalletAddress: wallet,
		DeployIfNeeded:       true,
		CreateOrDeriveAPIKey: true,
		ApproveTokens:        true,
		MaxApproval:          true,
		Relayer:              relayerClient,
		CLOB:                 clobClient,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Deployed || result.TokenApprovalsSubmitted {
		t.Fatalf("expected skipped deploy/approvals: %+v", result)
	}
	if !result.APIKeysCreated || !result.ReadyToTrade {
		t.Fatalf("expected API key recovered and ready: %+v", result)
	}
	if relayerClient.submittedCreate || relayerClient.submittedBatch {
		t.Fatalf("idempotent flow submitted relayer side effects")
	}
}

func TestResultRedactedDoesNotExposeSecrets(t *testing.T) {
	result := Result{APIKey: auth.APIKey{Key: "api-key-secret", Secret: "raw-secret-value", Passphrase: "raw-passphrase"}}
	redacted := result.Redacted()
	text := redacted.String()
	for _, secret := range []string{"api-key-secret", "raw-secret-value", "raw-passphrase"} {
		if strings.Contains(text, secret) {
			t.Fatalf("redacted output leaked %q: %s", secret, text)
		}
	}
	if !strings.Contains(text, "[REDACTED]") && !strings.Contains(text, "...") {
		t.Fatalf("redacted output did not redact: %s", text)
	}
}

func assertERC20Approve(t *testing.T, call map[string]string, token, spender string) {
	t.Helper()
	if !strings.EqualFold(call["target"], token) {
		t.Fatalf("target=%s want %s", call["target"], token)
	}
	data := strings.ToLower(call["data"])
	if !strings.HasPrefix(data, "0x095ea7b3") {
		t.Fatalf("approve selector missing: %s", data)
	}
	if !strings.Contains(data, strings.ToLower(strings.TrimPrefix(spender, "0x"))) {
		t.Fatalf("spender %s missing: %s", spender, data)
	}
	if !strings.HasSuffix(data, strings.Repeat("f", 64)) {
		t.Fatalf("maxUint256 missing: %s", data)
	}
}

type fakeRelayer struct {
	deployed        bool
	approvalsReady  bool
	nonce           string
	submittedCreate bool
	submittedBatch  bool
}

func (f *fakeRelayer) IsDeployed(ctx context.Context, ownerAddress string) (bool, error) {
	return f.deployed, nil
}
func (f *fakeRelayer) SubmitWalletCreate(ctx context.Context, ownerAddress string) (*relayer.RelayerTransaction, error) {
	f.submittedCreate = true
	f.deployed = true
	return &relayer.RelayerTransaction{TransactionID: "create-1", TransactionHash: "0xcreate", State: string(relayer.StateMined), Type: "WALLET-CREATE"}, nil
}
func (f *fakeRelayer) GetNonce(ctx context.Context, ownerAddress string) (string, error) {
	return f.nonce, nil
}
func (f *fakeRelayer) SubmitWalletBatch(ctx context.Context, ownerAddress, walletAddress, nonce, signature, deadline string, calls []relayer.DepositWalletCall) (*relayer.RelayerTransaction, error) {
	f.submittedBatch = true
	f.approvalsReady = true
	return &relayer.RelayerTransaction{TransactionID: "batch-1", TransactionHash: "0xbatch", State: string(relayer.StateMined), Type: "WALLET"}, nil
}
func (f *fakeRelayer) ApprovalsReady(ctx context.Context, walletAddress string) (bool, error) {
	return f.approvalsReady, nil
}

type fakeClob struct {
	existing bool
	called   bool
	key      auth.APIKey
}

func (f *fakeClob) CreateOrDeriveAPIKey(ctx context.Context, privateKey string) (auth.APIKey, error) {
	f.called = true
	if f.key.Key == "" {
		return auth.APIKey{}, errors.New("missing key")
	}
	return f.key, nil
}
func (f *fakeClob) VerifyTradingEnabled(ctx context.Context, privateKey string) (bool, error) {
	return f.existing || f.called, nil
}
