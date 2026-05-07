// Package risk provides per-trade caps, daily loss limits, and the
// circuit breaker that gates live order submission.
//
// Live commands consult risk before any signing or submission step.
// Read-only and paper modes do not call into this package. Limits are
// configured in internal/config and never derived from market data.
//
// This package is internal and not part of the polygolem public SDK.
package risk
