// Package polytypes holds the Polymarket protocol-level types shared
// across CLOB, Gamma, Data API, and stream clients.
//
// These types mirror the on-the-wire JSON shapes Polymarket returns.
// Keeping them in one place avoids per-client drift and lets paper-mode
// and live-mode reuse the same structures. There is no logic here —
// only types and small helpers.
//
// This package is internal and not part of the polygolem public SDK.
package polytypes
