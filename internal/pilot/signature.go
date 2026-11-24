package pilot

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type signedSnapshot struct {
	Format      string `json:"format"`
	Payload     string `json:"payload"`
	Digest      string `json:"digest"`
	PublicKey   string `json:"publicKey"`
	Fingerprint string `json:"fingerprint"`
	Signature   string `json:"signature"`
}

func signingKey(dir string) (ed25519.PrivateKey, error) {
	path := filepath.Join(dir, "demo-signing.key")
	seed, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		seed = make([]byte, ed25519.SeedSize)
		if _, err = rand.Read(seed); err != nil {
			return nil, err
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			seed, err = os.ReadFile(path)
		} else if err == nil {
			_, err = file.Write(seed)
			if err == nil {
				err = file.Sync()
			}
			closeErr := file.Close()
			if err == nil {
				err = closeErr
			}
		}
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid local signing seed")
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

func signDemoSnapshot(dir, project string, events []demoEvent) (signedSnapshot, error) {
	return signSnapshot(dir, project, events, true)
}

func signSnapshot(dir, project string, events []demoEvent, fictional bool) (signedSnapshot, error) {
	rows := []demoEvent{}
	for _, event := range events {
		if event.Project == project {
			rows = append(rows, event)
		}
	}
	if len(rows) == 0 {
		return signedSnapshot{}, fmt.Errorf("project unavailable")
	}
	payload, err := json.Marshal(struct {
		Fictional  bool        `json:"fictional"`
		ExportedAt string      `json:"exportedAt"`
		Events     []demoEvent `json:"events"`
	}{fictional, time.Now().UTC().Format(time.RFC3339Nano), rows})
	if err != nil {
		return signedSnapshot{}, err
	}
	key, err := signingKey(dir)
	if err != nil {
		return signedSnapshot{}, err
	}
	pub := key.Public().(ed25519.PublicKey)
	digest := sha256.Sum256(payload)
	fingerprint := sha256.Sum256(pub)
	format := "nirikshan-workspace-v1"
	if fictional {
		format = "nirikshan-demo-v2"
	}
	return signedSnapshot{
		Format: format, Payload: string(payload), Digest: hex.EncodeToString(digest[:]),
		PublicKey: base64.StdEncoding.EncodeToString(pub), Fingerprint: hex.EncodeToString(fingerprint[:]),
		Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(key, payload)),
	}, nil
}
