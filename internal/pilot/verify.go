package pilot

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type VerifiedEvidence struct {
	Project     string
	Events      int
	Photos      int
	Fingerprint string
	Trusted     bool
}

// VerifyEvidence checks a signed export independently of the server and its private key.
func VerifyEvidence(reader io.Reader, expectedFingerprint string) (VerifiedEvidence, error) {
	const limit = 50 << 20
	content, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil || len(content) > limit {
		return VerifiedEvidence{}, fmt.Errorf("could not read bundle or bundle exceeds 50 MiB")
	}
	var bundle struct {
		signedSnapshot
		Photos map[string]string `json:"photos"`
	}
	if err := json.Unmarshal(content, &bundle); err != nil || (bundle.Format != "nirikshan-demo-v2" && bundle.Format != "nirikshan-workspace-v1") {
		return VerifiedEvidence{}, fmt.Errorf("expected a signed Nirikshan evidence bundle")
	}
	payload := []byte(bundle.Payload)
	digest := sha256.Sum256(payload)
	if bundle.Digest != hex.EncodeToString(digest[:]) {
		return VerifiedEvidence{}, fmt.Errorf("record digest mismatch")
	}
	publicKey, err := base64.StdEncoding.DecodeString(bundle.PublicKey)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return VerifiedEvidence{}, fmt.Errorf("invalid signing key")
	}
	fingerprint := sha256.Sum256(publicKey)
	fingerprintHex := hex.EncodeToString(fingerprint[:])
	if bundle.Fingerprint != fingerprintHex {
		return VerifiedEvidence{}, fmt.Errorf("signing fingerprint mismatch")
	}
	expectedFingerprint = strings.ToLower(strings.TrimSpace(expectedFingerprint))
	if expectedFingerprint != "" && expectedFingerprint != fingerprintHex {
		return VerifiedEvidence{}, fmt.Errorf("signer differs from trusted fingerprint")
	}
	signature, err := base64.StdEncoding.DecodeString(bundle.Signature)
	if err != nil || !ed25519.Verify(publicKey, payload, signature) {
		return VerifiedEvidence{}, fmt.Errorf("invalid record signature")
	}
	var snapshot struct {
		Fictional bool        `json:"fictional"`
		Events    []demoEvent `json:"events"`
	}
	if err := json.Unmarshal(payload, &snapshot); err != nil || snapshot.Fictional != (bundle.Format == "nirikshan-demo-v2") || len(snapshot.Events) == 0 {
		return VerifiedEvidence{}, fmt.Errorf("invalid project snapshot")
	}
	project := snapshot.Events[0].Project
	photos := map[string]bool{}
	for _, event := range snapshot.Events {
		if event.Project != project {
			return VerifiedEvidence{}, fmt.Errorf("inconsistent project identifiers")
		}
		if event.PhotoHash == "" || photos[event.PhotoHash] {
			continue
		}
		photo, err := base64.StdEncoding.DecodeString(bundle.Photos[event.PhotoHash])
		digest := sha256.Sum256(photo)
		if err != nil || hex.EncodeToString(digest[:]) != event.PhotoHash {
			return VerifiedEvidence{}, fmt.Errorf("missing or modified photo %s", event.PhotoHash)
		}
		photos[event.PhotoHash] = true
	}
	if len(bundle.Photos) != len(photos) {
		return VerifiedEvidence{}, fmt.Errorf("bundle includes unreferenced photos")
	}
	return VerifiedEvidence{project, len(snapshot.Events), len(photos), fingerprintHex, expectedFingerprint != ""}, nil
}
