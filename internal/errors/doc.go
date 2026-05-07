// Package errors provides structured error types and code helpers used
// across polygolem clients and command handlers.
//
// Each error carries a stable code suitable for surfacing in the JSON
// envelope rendered by internal/output. Wrap protocol errors with the
// helpers here rather than reaching for fmt.Errorf so downstream
// consumers can switch on Code.
//
// This package is internal and not part of the polygolem public SDK.
package errors
