// Package preflight runs local and remote readiness probes before
// polygolem performs any state-changing operation.
//
// A Probe is a context-aware function returning an error. Probes cover
// builder credentials, deposit-wallet status, RPC reachability, and
// relayer health. The aggregate result gates entry into live mode.
//
// This package is internal and not part of the polygolem public SDK.
package preflight
