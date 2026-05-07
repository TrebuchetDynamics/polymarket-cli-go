// Package gamma is the typed Gamma HTTP client used internally by polygolem
// — markets, events, search, tags, series, sports, comments, and profiles.
//
// Gamma is Polymarket's read-only metadata API. This client wraps it with
// retry, rate-limiting, and structured types from internal/polytypes. It
// performs no signing and never mutates state. Start with Client and
// Client.Search for orientation.
//
// This package is internal and not part of the polygolem public SDK.
// External consumers should use pkg/gamma instead.
package gamma
