// Package modes parses and gates the polygolem operating modes —
// read-only, paper, and live.
//
// Mode selection comes from configuration and CLI flags. Read-only is the
// default and never reaches authenticated mutation endpoints. Paper stays
// local. Live requires preflight, risk, and funding gates to pass before
// any signed call goes out.
//
// This package is internal and not part of the polygolem public SDK.
package modes
