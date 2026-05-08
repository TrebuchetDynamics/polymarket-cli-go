package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteErrorJSONUsesStableEnvelope(t *testing.T) {
	var buf bytes.Buffer

	err := WriteError(&buf, FormatJSON, Error{
		Code:    "live_gate_failed",
		Message: "live trading requires --confirm-live",
		Details: map[string]string{"gate": "cli_confirmation"},
	})
	if err != nil {
		t.Fatalf("WriteError returned error: %v", err)
	}

	var got struct {
		OK      bool   `json:"ok"`
		Version string `json:"version"`
		Error   Error  `json:"error"`
		Meta    Meta   `json:"meta"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("error output is not JSON: %v\n%s", err, buf.String())
	}
	if got.OK {
		t.Fatalf("ok=true, want false: %#v", got)
	}
	if got.Version != ContractVersion {
		t.Fatalf("version=%q, want %q", got.Version, ContractVersion)
	}
	if got.Error.Code != "live_gate_failed" {
		t.Fatalf("unexpected code: %#v", got)
	}
	if got.Error.Message != "live trading requires --confirm-live" {
		t.Fatalf("unexpected message: %#v", got)
	}
	if got.Meta.TS == "" {
		t.Fatalf("meta timestamp missing: %#v", got)
	}
}
