// Package paper holds local-only paper-trading state — positions, fills,
// and persisted snapshots.
//
// Paper state lives entirely on disk and never reaches an authenticated
// Polymarket endpoint. The paper executor in internal/execution writes
// here; live mode does not touch this package. Useful for replay, edge
// validation, and offline development.
//
// This package is internal and not part of the polygolem public SDK.
package paper
