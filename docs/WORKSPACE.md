# Nirikshan department workspace

The workspace is a single-department application with locally provisioned accounts. It starts empty. The fictional demo, public Udaipur tender atlas and authenticated records use separate data.

## Provision accounts

Run each command from the repository root with the same `DATA_DIR` that the service uses. Passwords come from standard input. Choose a different password for each account. On Windows, use PowerShell 7 and enter a password when prompted.

```powershell
Read-Host -MaskInput "Department password" | go run . account-add department government
Read-Host -MaskInput "Contractor password" | go run . account-add contractor contractor udaipur-road
Read-Host -MaskInput "Engineer password" | go run . account-add engineer engineer udaipur-road
Read-Host -MaskInput "Citizen password" | go run . account-add resident citizen udaipur-road
go run . -workspace
```

On Bash, use `read -rs password` to read a password, then `printf '%s\\n' "$password" | go run . account-add department government` and `unset password`. Repeat for each account and project assignment. Usernames and project identifiers may use up to 64 lowercase letters, digits or hyphens and start with a letter or digit.

Open [the local workspace](http://localhost:8080/lifecycle.html?mode=workspace). Sign in as the department and record a sanction with identifier `udaipur-road`. Government accounts cover every project in the installation. Contractor, engineer and citizen accounts receive only the project IDs provided at provisioning.

## Follow a project

1. The department enters sanction, dates, scope and route endpoints. Select **Tender published** to import a UTF-8 notice or paste copied text. Use labeled lines such as `Tender ID: ...`, `Work Title: ...`, `Tender Value: ₹1,97,55,000` and `Official URL: https://...`. Confirm the extracted amount, reference and source before saving. Scanned PDFs require manual transcription or OCR before pasting.
2. Assigned contractors record quotations. The department can compare them and awards one recorded quotation, then records a work order and approved work items. The awarded contractor can record construction, progress, deliveries and bills.
3. The engineer records measurements, inspection and any quality tests. The department can record a payment up to the remaining certified amount. Financial signals identify records worth reviewing.
4. Assigned citizens can submit photos and optional approximate location, view their own reports and reopen resolved issues. The department can acknowledge and resolve reports.
5. Select a project in the portfolio, then download the audit CSV or signed evidence. Exports reflect the signed-in account's view. The audit CSV records named actors, amounts in integer paise and evidence references. Verify the evidence bundle with `go run ./cmd/verify -fingerprint SHA256 FILE.json`, using the fingerprint obtained independently.

## Operations

The workspace binds to loopback. For an internal HTTPS pilot, terminate TLS at a local reverse proxy and set `NIRIKSHAN_WORKSPACE_HTTPS=1` for secure session cookies. Use one server process per data directory on a local filesystem. Back up `DATA_DIR/workspace` as a unit, preferably while stopped. It contains account password verifiers, project histories, photos and the signing key. Keep audit exports private and limit distribution to authorized recipients.

Account provisioning is operator-managed. To revoke an account, stop the server, remove its account file under `DATA_DIR/workspace/accounts`, and restart. Startup clears existing sessions. Reprovision with the same username if a password or assignment changes. Keep the original signing key to maintain the fingerprint across recovery. An installation signature attests to the exported bytes and signing identity. The department remains responsible for the underlying records.

Official staff identity verification, external grievance integration and production hosting require the operating department.
