# Nirikshan demonstration

## Start the app

Run `go run . -demo` from the repository root and open http://localhost:8080.

Use a fresh data directory when preparing a new walkthrough. On Windows:

```powershell
$env:DATA_DIR = "data/presentation"
go run . -demo
```

The public atlas contains five sourced tender notices. The two demonstration projects are fictional and are marked as such throughout the lifecycle pages.

## An eight-minute presentation

| Time | Show | What to do |
| --- | --- | --- |
| 0:00 | The public problem | Open the public atlas, select Boramcha bridge, and open its official tender. Show the distinction between advertised value and unavailable payment records. |
| 1:00 | The lifecycle | Open the interactive demo. Select the Aravalli demonstration bridge. Tap the sanction, tender, award and construction stages. The school project illustrates the entire lifecycle through maintenance. |
| 2:00 | Financial transparency | Open finances. The bridge's illustrative concrete bill lists 120 m³, and the engineer certified 100 m³ at ₹6,000/m³. ₹1.2 lakh awaits certification. |
| 3:00 | Quality evidence | Open Quality. Inspect the fictional sample LAB-DEMO-17 comparing a 28 MPa measurement against the recorded 30 MPa minimum. |
| 4:00 | Citizen participation | Report an issue, add an optional photo and location, and save the receipt. The department desk will now show the report. |
| 5:00 | Department response | Use the report's acknowledgement button, record a response, then record a resolution. Reopen the issue if further review is needed. |
| 6:30 | Evidence integrity | Export the signed bundle, save the fingerprint and verify the export against it. Modify a copy of the signed payload and show that verification fails. |
| 7:30 | Missing information | Return to the real atlas and open the records request builder. Choose a Udaipur tender and prepare an English or Hindi draft. |

## Create a completely new project

Use the department desk to record a sanction with a unique identifier, scope, planned dates and amount. Enter the tender next. Switch to the contractor desk to submit a quotation, and return to the department desk to select that quotation for the award. Continue with the work order and work items.

The contractor can record deliveries, progress and bills. The engineer records measurements and quality results. The department records payments against certified bills. The tracker validates required stages and rejects payments above the certified unpaid amount.

## Independently verify the evidence

Download the signed project bundle from the verification panel. It contains a signed JSON snapshot and the referenced photo bytes. Use a copy of the fingerprint recorded independently of the file.

```sh
go run ./cmd/verify -fingerprint YOUR_TRUSTED_FINGERPRINT nirikshan-demo-evidence.json
```

A valid signature identifies the local installation. It establishes the integrity of the signed records and referenced photos. The recorded information and physical work still require independent verification.

## Demonstration boundary

Fictional records and roles provide an end-to-end product experience. The five public tender notices remain separately sourced. The application can be evaluated locally without spending money or contacting a government office. Verified officer identities, official data feeds, genuine department acknowledgement of grievances and independently published ledger checkpoints are the next real-pilot milestones.
