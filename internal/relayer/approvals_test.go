package relayer

import (
	"strings"
	"testing"
)

func TestBuildAdapterApprovalCallsCalldata(t *testing.T) {
	calls := BuildAdapterApprovalCalls()
	if len(calls) != 4 {
		t.Fatalf("len=%d want 4", len(calls))
	}

	// Calls 0-1: CtfCollateralAdapter
	assertApprove(t, "call0", calls[0], pusdAddress, ctfCollateralAdapter)
	assertSetApprovalForAll(t, "call1", calls[1], ctfAddress, ctfCollateralAdapter)

	// Calls 2-3: NegRiskCtfCollateralAdapter
	assertApprove(t, "call2", calls[2], pusdAddress, negRiskCtfCollateralAdapter)
	assertSetApprovalForAll(t, "call3", calls[3], ctfAddress, negRiskCtfCollateralAdapter)
}

func TestBuildAdapterApprovalCallsIdempotent(t *testing.T) {
	a := BuildAdapterApprovalCalls()
	b := BuildAdapterApprovalCalls()
	if len(a) != len(b) {
		t.Fatalf("len mismatch a=%d b=%d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("call %d differs", i)
		}
	}
}

func assertApprove(t *testing.T, label string, call DepositWalletCall, expectToken, expectSpender string) {
	t.Helper()
	if !strings.EqualFold(call.Target, expectToken) {
		t.Errorf("%s target=%s want %s", label, call.Target, expectToken)
	}
	data := strings.ToLower(call.Data)
	if !strings.HasPrefix(data, "0x"+erc20ApproveSelector) {
		t.Errorf("%s data does not start with approve selector: %s", label, data[:10])
	}
	wantSpender := strings.ToLower(strings.TrimPrefix(expectSpender, "0x"))
	if !strings.Contains(data, wantSpender) {
		t.Errorf("%s spender %s not encoded in calldata: %s", label, wantSpender, data)
	}
	if !strings.HasSuffix(data, maxUint256) {
		t.Errorf("%s amount is not MaxUint256: %s", label, data)
	}
}

func assertSetApprovalForAll(t *testing.T, label string, call DepositWalletCall, expectCTF, expectOperator string) {
	t.Helper()
	if !strings.EqualFold(call.Target, expectCTF) {
		t.Errorf("%s target=%s want %s", label, call.Target, expectCTF)
	}
	data := strings.ToLower(call.Data)
	if !strings.HasPrefix(data, "0x"+erc1155SetApprovalForAllSel) {
		t.Errorf("%s data does not start with setApprovalForAll selector: %s", label, data[:10])
	}
	wantOp := strings.ToLower(strings.TrimPrefix(expectOperator, "0x"))
	if !strings.Contains(data, wantOp) {
		t.Errorf("%s operator %s not encoded in calldata: %s", label, wantOp, data)
	}
	if !strings.HasSuffix(data, "0000000000000000000000000000000000000000000000000000000000000001") {
		t.Errorf("%s approved=true not encoded: %s", label, data)
	}
}
