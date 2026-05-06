package modes

import "testing"

func TestDefaultModeIsReadOnly(t *testing.T) {
	mode, err := Parse("")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if mode != ReadOnly {
		t.Fatalf("mode = %q, want %q", mode, ReadOnly)
	}
}

func TestLiveModeRequiresAllGates(t *testing.T) {
	result := ValidateLiveGates(LiveGateInput{
		EnvEnabled:    true,
		ConfigEnabled: true,
		ConfirmLive:   false,
		PreflightOK:   true,
	})
	if result.Allowed {
		t.Fatal("live mode allowed without CLI confirmation")
	}
	if result.Failures[0].Code != "cli_confirmation_required" {
		t.Fatalf("unexpected failures: %#v", result.Failures)
	}
}
