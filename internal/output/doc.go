// Package output renders command results as either tables or stable JSON
// envelopes and emits structured error responses.
//
// Tables are designed for humans; the JSON envelope is the contract every
// command handler honors when --json is set. Handlers should call into
// this package rather than printing directly so the envelope shape stays
// stable.
//
// This package is internal and not part of the polygolem public SDK.
package output
