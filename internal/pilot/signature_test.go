package pilot

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestSignedSnapshot(t *testing.T) {
	dir := t.TempDir()
	first, err := signDemoSnapshot(dir, "demo-bridge", demoSeed())
	if err != nil {
		t.Fatal(err)
	}
	second, err := signDemoSnapshot(dir, "demo-bridge", demoSeed())
	if err != nil {
		t.Fatal(err)
	}
	if first.PublicKey != second.PublicKey || first.Fingerprint != second.Fingerprint {
		t.Fatal("signer identity changed")
	}
	pub, err := base64.StdEncoding.DecodeString(first.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := base64.StdEncoding.DecodeString(first.Signature)
	if err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(pub, []byte(first.Payload), sig) {
		t.Fatal("signature failed")
	}
	if ed25519.Verify(pub, []byte(first.Payload+" "), sig) {
		t.Fatal("altered payload accepted")
	}
	if _, err := signDemoSnapshot(dir, "missing", demoSeed()); err == nil {
		t.Fatal("unknown project signed")
	}
	bundle, err := json.Marshal(struct {
		signedSnapshot
		Photos map[string]string `json:"photos"`
	}{first, map[string]string{}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := VerifyEvidence(bytes.NewReader(bundle), first.Fingerprint)
	if err != nil || !result.Trusted || result.Project != "demo-bridge" {
		t.Fatalf("offline verification: %+v, %v", result, err)
	}
	if _, err := VerifyEvidence(bytes.NewReader(bundle), "wrong-fingerprint"); err == nil {
		t.Fatal("wrong fingerprint accepted")
	}
}
