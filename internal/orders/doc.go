// Package orders defines OrderIntent, the fluent builder, validation
// rules, and order lifecycle states used by both paper and live executors.
//
// An OrderIntent is the protocol-agnostic shape an executor accepts.
// Construction goes through Builder so invariants (size, side, market,
// signature type) are checked before any network call. Lifecycle states
// are explicit so paper and live can share the same state machine.
//
// This package is internal and not part of the polygolem public SDK.
package orders
