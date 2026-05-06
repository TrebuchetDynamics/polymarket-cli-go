package modes

import "fmt"

type Mode string

const (
	ReadOnly Mode = "read-only"
	Paper    Mode = "paper"
	Live     Mode = "live"
)

type Failure struct {
	Code    string
	Message string
}

type LiveGateInput struct {
	EnvEnabled    bool
	ConfigEnabled bool
	ConfirmLive   bool
	PreflightOK   bool
}

type LiveGateResult struct {
	Allowed  bool
	Failures []Failure
}

func Parse(value string) (Mode, error) {
	switch value {
	case "", string(ReadOnly):
		return ReadOnly, nil
	case string(Paper):
		return Paper, nil
	case string(Live):
		return Live, nil
	default:
		return "", fmt.Errorf("unknown mode %q", value)
	}
}

func ValidateLiveGates(input LiveGateInput) LiveGateResult {
	var failures []Failure
	if !input.EnvEnabled {
		failures = append(failures, Failure{Code: "env_gate_required", Message: "POLYMARKET_LIVE_PROFILE must be on"})
	}
	if !input.ConfigEnabled {
		failures = append(failures, Failure{Code: "config_gate_required", Message: "live_trading_enabled must be true"})
	}
	if !input.ConfirmLive {
		failures = append(failures, Failure{Code: "cli_confirmation_required", Message: "--confirm-live is required"})
	}
	if !input.PreflightOK {
		failures = append(failures, Failure{Code: "preflight_required", Message: "preflight must pass"})
	}
	return LiveGateResult{Allowed: len(failures) == 0, Failures: failures}
}
