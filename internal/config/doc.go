// Package config loads polygolem configuration via Viper — defaults,
// environment binding, file overrides, validation, and credential
// redaction.
//
// Every entry point reads config through Load. Builder credentials and
// private keys are redacted at load time so no downstream logger or JSON
// emitter ever sees the plaintext value.
//
// This package is internal and not part of the polygolem public SDK.
package config
