// Package transport is the shared HTTP client layer — retry, rate limiting,
// circuit breaking, and credential redaction.
//
// Every protocol client (gamma, clob, dataapi, relayer, bridge) sits on
// top of this. Configure with DefaultConfig and inject via Client. The
// redactor is wired in by internal/config so credentials never reach
// stdout or logs.
//
// This package is internal and not part of the polygolem public SDK.
package transport
