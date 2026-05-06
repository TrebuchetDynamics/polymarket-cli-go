package preflight

import (
	"context"
	"errors"
	"testing"
)

func TestRunReportsProbeFailures(t *testing.T) {
	checks := []Check{
		{Name: "gamma", Probe: func(context.Context) error { return nil }},
		{Name: "clob", Probe: func(context.Context) error { return errors.New("503") }},
	}
	result := Run(context.Background(), checks)
	if result.OK {
		t.Fatal("preflight should fail when a probe fails")
	}
	if result.Checks[1].Status != "fail" {
		t.Fatalf("second status = %q", result.Checks[1].Status)
	}
}
