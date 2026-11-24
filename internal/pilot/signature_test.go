package pilot

import (
	"crypto/ed25519"
	"encoding/base64"
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
}
