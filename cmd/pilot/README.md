# Pilot service

Run `go run ./cmd/pilot` from the repository root. The process reads `web/projects.json` at startup, serves the web directory, and listens on PORT (default 8080). WEB_DIR overrides the web path. DATA_DIR defaults to data. Use one server instance with a durable filesystem supporting hard links. Back up the event directory and test restoration. Keep it outside the public web directory.

Set REVIEW_TOKEN to a random private reviewer credential. Without it the service is read-only. Use HTTPS for hosted access. GET /api/projects returns the public project catalog. GET /api/evidence?project=id returns approved evidence. POST /api/evidence accepts JSON containing project, kind, title, YYYY-MM-DD date, HTTPS url, optional reference and optional nonnegative integer amountPaise for bill/payment/award. Observation URLs may be omitted. Successful submissions receive HTTP 202 with id and status pending.

GET /api/review returns a pending array to a reviewer. POST /api/review accepts id, decision (approve or reject), and a required reason. Both use Authorization: Bearer <REVIEW_TOKEN>. A contributor can look up their private reference with GET /api/receipts/{id} to retrieve the status pending, approve, or reject. Health checks use GET /healthz.

Each submission or decision creates an immutable-by-convention numbered event, linked to the previous SHA-256 digest. Startup validates the chain and transitions. Off-host backups or independent digest publication are required to detect wholesale rewriting or suffix removal. Review authentication is a shared pilot credential. A department deployment needs named reviewer accounts and supervision. The submission limiter keys on the immediate peer socket address. A reverse proxy may group clients under the same limit.
