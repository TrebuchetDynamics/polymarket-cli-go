// Package marketdiscovery provides high-level market discovery by
// joining Gamma metadata with CLOB tick-size and orderbook details.
//
// Wraps internal/gamma and internal/clob so command handlers (`discover`,
// `discover enrich`, `discover market`) can return one denormalized view
// instead of stitching Gamma plus CLOB calls per command. Read-only;
// safe in every mode.
//
// This package is internal and not part of the polygolem public SDK.
package marketdiscovery
