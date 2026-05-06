package preflight

import "context"

type Probe func(context.Context) error

type Check struct {
	Name  string
	Probe Probe
}

type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Result struct {
	OK     bool          `json:"ok"`
	Checks []CheckResult `json:"checks"`
}

func Run(ctx context.Context, checks []Check) Result {
	result := Result{OK: true}
	for _, check := range checks {
		err := check.Probe(ctx)
		checkResult := CheckResult{Name: check.Name, Status: "pass"}
		if err != nil {
			result.OK = false
			checkResult.Status = "fail"
			checkResult.Message = err.Error()
		}
		result.Checks = append(result.Checks, checkResult)
	}
	return result
}
