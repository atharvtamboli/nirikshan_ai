# Pilot release readiness

## Release scope

Showcase and controlled local departmental pilot. Public atlas and fictional showcase are separate from the authenticated workspace.

- Local launch: `go run . -demo`; authenticated workflow: `go run . -workspace` after named accounts are provisioned. See [workspace guide](WORKSPACE.md).
- Go 1.24.1+ and Node.js for tests. No paid services required.
- Confirm the sanction → quotation → award → work order → inspection → certified payment flow. Use synthetic records.
- Verify a signed export against an independently recorded signing fingerprint; review CSV attribution.
- Preserve the data directory and signing key securely; rehearse restore before a pilot.

## Validation

Reviewed main commit 1309f65ff8049fd0fdc63dee366d894a6f938951. Local Go tests, vet, five JavaScript tests and a Windows build passed. The authenticated-workspace PR's GitHub Actions checks passed, including its Windows browser workflow.

## Gates before deployment

Verified officer identity and departmental authority, authenticated official project records, consent and moderation for citizen reports, independently anchored blockchain receipts (Ethereum integration remains pending), durable multi-user production storage, backups, access review, monitoring and approved HTTPS hosting. Obtain a real department pilot before any production release. A signed record establishes integrity and attribution, not proof that the underlying claim is true.
