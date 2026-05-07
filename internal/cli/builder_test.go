package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuilderOnboardCommandRegistered confirms the command is wired into
// the root tree. Drives the wiring change in root.go.
func TestBuilderOnboardCommandRegistered(t *testing.T) {
	root := NewRootCommand(Options{Version: "test", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	builderCmd, _, err := root.Find([]string{"builder", "onboard"})
	if err != nil {
		t.Fatalf("builder onboard not found: %v", err)
	}
	if builderCmd.Use != "onboard" {
		t.Fatalf("expected Use=onboard, got %q", builderCmd.Use)
	}
}

// TestValidateBuilderCredentialFormatAcceptsValidShape pins the loose
// shape we accept. The real Polymarket creds are: API key in UUID v4
// shape, secret base64-encoded (typically 32–64 bytes raw → 44+ chars),
// passphrase opaque non-empty.
func TestValidateBuilderCredentialFormatAcceptsValidShape(t *testing.T) {
	err := validateBuilderCredentialFormat(
		"f3399c2e-aaaa-4bbb-8ccc-deadbeefeada",
		"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU2Nzg5", // base64 of 36 bytes
		"some-opaque-passphrase",
	)
	if err != nil {
		t.Fatalf("expected valid creds to pass, got %v", err)
	}
}

func TestValidateBuilderCredentialFormatRejectsBadAPIKey(t *testing.T) {
	err := validateBuilderCredentialFormat(
		"not-a-uuid",
		"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU2Nzg5",
		"opaque",
	)
	if err == nil {
		t.Fatal("expected non-UUID api key to be rejected")
	}
	if !strings.Contains(err.Error(), "api key") {
		t.Fatalf("expected error to mention api key, got %v", err)
	}
}

func TestValidateBuilderCredentialFormatRejectsNonBase64Secret(t *testing.T) {
	err := validateBuilderCredentialFormat(
		"f3399c2e-aaaa-4bbb-8ccc-deadbeefeada",
		"this-has-bad-padding!",
		"opaque",
	)
	if err == nil {
		t.Fatal("expected non-base64 secret to be rejected")
	}
	if !strings.Contains(err.Error(), "secret") {
		t.Fatalf("expected error to mention secret, got %v", err)
	}
}

func TestValidateBuilderCredentialFormatRejectsEmptyPassphrase(t *testing.T) {
	err := validateBuilderCredentialFormat(
		"f3399c2e-aaaa-4bbb-8ccc-deadbeefeada",
		"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU2Nzg5",
		"",
	)
	if err == nil {
		t.Fatal("expected empty passphrase to be rejected")
	}
	if !strings.Contains(err.Error(), "passphrase") {
		t.Fatalf("expected error to mention passphrase, got %v", err)
	}
}

// TestPersistBuilderCredentialsWritesPosixMode0600 — the resulting env
// file must be operator-only readable so secrets don't leak via shared
// filesystems or cloud sync.
func TestPersistBuilderCredentialsWritesPosixMode0600(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".env.builder")
	creds := builderCreds{
		Key:        "f3399c2e-aaaa-4bbb-8ccc-deadbeefeada",
		Secret:     "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU2Nzg5",
		Passphrase: "opaque",
	}
	if err := persistBuilderCredentials(target, creds, false); err != nil {
		t.Fatalf("persist returned error: %v", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected mode 0600, got %o", info.Mode().Perm())
	}
}

// TestPersistBuilderCredentialsContainsAllThreeVars — the written file
// must export the three POLYMARKET_BUILDER_* vars in shell-sourceable
// form.
func TestPersistBuilderCredentialsContainsAllThreeVars(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".env.builder")
	creds := builderCreds{
		Key:        "f3399c2e-aaaa-4bbb-8ccc-deadbeefeada",
		Secret:     "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU2Nzg5",
		Passphrase: "opaque-pass-with-spaces and special$chars",
	}
	if err := persistBuilderCredentials(target, creds, false); err != nil {
		t.Fatalf("persist: %v", err)
	}
	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	got := string(body)
	for _, want := range []string{
		"POLYMARKET_BUILDER_API_KEY=",
		"POLYMARKET_BUILDER_SECRET=",
		"POLYMARKET_BUILDER_PASSPHRASE=",
		creds.Key,
		creds.Secret,
		creds.Passphrase,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("env file missing %q\nbody: %s", want, got)
		}
	}
}

// TestPersistBuilderCredentialsRefusesOverwriteWithoutForce — the
// command must not silently clobber existing builder creds.
func TestPersistBuilderCredentialsRefusesOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".env.builder")
	creds := builderCreds{
		Key:        "f3399c2e-aaaa-4bbb-8ccc-deadbeefeada",
		Secret:     "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU2Nzg5",
		Passphrase: "v1",
	}
	if err := persistBuilderCredentials(target, creds, false); err != nil {
		t.Fatalf("first write: %v", err)
	}
	creds.Passphrase = "v2"
	err := persistBuilderCredentials(target, creds, false)
	if err == nil {
		t.Fatal("expected refuse to overwrite without force")
	}
	if !strings.Contains(err.Error(), "exists") && !strings.Contains(err.Error(), "force") {
		t.Fatalf("expected error to mention existing/force, got %v", err)
	}
	body, _ := os.ReadFile(target)
	if !strings.Contains(string(body), "v1") {
		t.Fatal("first creds should not have been overwritten")
	}
}

// TestPersistBuilderCredentialsForceOverwrites — with force=true, an
// existing file is replaced.
func TestPersistBuilderCredentialsForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".env.builder")
	creds := builderCreds{Key: "f3399c2e-aaaa-4bbb-8ccc-deadbeefeada", Secret: "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU2Nzg5", Passphrase: "v1"}
	_ = persistBuilderCredentials(target, creds, false)
	creds.Passphrase = "v2"
	if err := persistBuilderCredentials(target, creds, true); err != nil {
		t.Fatalf("force write: %v", err)
	}
	body, _ := os.ReadFile(target)
	if !strings.Contains(string(body), "v2") {
		t.Fatal("force should have overwritten")
	}
	if strings.Contains(string(body), "v1") {
		t.Fatal("v1 leaked into the new file")
	}
}

// TestBuilderOnboardJSONOutputShape — the success envelope must have
// the expected shape so calling shell scripts can rely on it.
func TestBuilderOnboardJSONOutputShape(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".env.builder")
	creds := builderCreds{
		Key:        "f3399c2e-aaaa-4bbb-8ccc-deadbeefeada",
		Secret:     "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU2Nzg5",
		Passphrase: "opaque",
	}
	if err := persistBuilderCredentials(target, creds, false); err != nil {
		t.Fatalf("persist: %v", err)
	}

	out := builderOnboardResult{
		WroteTo:    target,
		Validated:  false,
		Permission: "0600",
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	for _, want := range []string{`"wrote_to"`, `"validated"`, `"permission"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("envelope missing field %q in %s", want, got)
		}
	}
}
