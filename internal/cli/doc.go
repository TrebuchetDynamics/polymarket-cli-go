// Package cli builds the polygolem Cobra command tree and wires command
// handlers to typed protocol, execution, and safety packages.
//
// Every command lives here and delegates to a typed package — handlers do
// not contain protocol logic. The default invocation enters read-only
// mode; live commands require an explicit signature type and gate pass.
// Start with NewRootCmd for orientation.
//
// This package is internal and not part of the polygolem public SDK.
package cli
