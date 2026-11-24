package main

import (
	"flag"
	"fmt"
	"os"

	"nirikshan.ai/pilot/internal/pilot"
)

func main() {
	fingerprint := flag.String("fingerprint", "", "independently trusted signing fingerprint")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/verify [-fingerprint SHA256] evidence.json")
		os.Exit(2)
	}
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()
	result, err := pilot.VerifyEvidence(file, *fingerprint)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Verification failed:", err)
		os.Exit(1)
	}
	fmt.Printf("Signature and evidence digests valid\nProject: %s\nEvents: %d\nPhotos: %d\nFingerprint: %s\n", result.Project, result.Events, result.Photos, result.Fingerprint)
	if result.Trusted {
		fmt.Println("Signer matches trusted fingerprint.")
	} else {
		fmt.Println("Independently confirm the fingerprint to establish signer identity.")
	}
}
