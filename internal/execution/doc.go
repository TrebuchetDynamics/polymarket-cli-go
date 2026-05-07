// Package execution defines the executor interface and ships the
// paper-mode implementation. A live executor satisfies the same contract.
//
// Executors take a validated OrderIntent and return a typed result.
// Paper executors update internal/paper state; live executors call
// internal/clob and internal/relayer behind the live-mode gate. Handlers
// depend on the interface, never on a concrete executor.
//
// This package is internal and not part of the polygolem public SDK.
package execution
