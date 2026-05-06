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

	var got map[string]map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("error output is not JSON: %v\n%s", err, buf.String())
	}
	if got["error"]["code"] != "live_gate_failed" {
		t.Fatalf("unexpected code: %#v", got)
	}
	if got["error"]["message"] != "live trading requires --confirm-live" {
		t.Fatalf("unexpected message: %#v", got)
	}
}
