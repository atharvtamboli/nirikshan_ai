# 🏗️ Nirikshan | Autonomous Infrastructure Ledger

**Nirikshan** is a Go-based blockchain application designed to bring transparency and accountability to public infrastructure projects.

The system records project proposals, updates, spending information, and audit activity on a tamper-evident blockchain ledger. Instead of silently modifying an existing record, every change becomes part of the project's historical record.

The current version focuses on the core blockchain implementation, Proof-of-Work mining, project tracking, persistent storage, and a visual blockchain dashboard.

---

## 🌟 What Nirikshan Does

Nirikshan provides a blockchain-backed system for tracking public infrastructure projects.

### Current capabilities

* ⛓️ **Blockchain Ledger**
  Maintains a chain of blocks linked through cryptographic hashes.

* ⛏️ **Proof-of-Work**
  Blocks are mined using a configurable-style PoW mechanism to demonstrate computational consensus.

* 🔐 **SHA-256 Hashing**
  Blocks use SHA-256 hashes to maintain the integrity of the chain.

* 📋 **Project Tracking**
  Infrastructure projects can be created and updated through the application.

* 💰 **Financial & Audit Records**
  Project spending and updates are recorded as new blockchain entries.

* 💾 **Ledger Persistence**
  The blockchain is stored in `ledger_backup.json` so the ledger can survive application restarts.

* 🌐 **REST API**
  The Go backend exposes APIs for interacting with the blockchain and project data.

* 🎨 **3D Blockchain Visualization**
  Three.js is used to visualize the blockchain through an interactive dashboard.

* 🖥️ **Modern Web Interface**
  The frontend uses HTML, Tailwind CSS, and Three.js with a glassmorphism-inspired interface.

---

## 🛠️ Technology Stack

| Layer         | Technology           |
| ------------- | -------------------- |
| Backend       | Go                   |
| Web Framework | Gin                  |
| Frontend      | HTML5                |
| Styling       | Tailwind CSS         |
| Visualization | Three.js             |
| Hashing       | SHA-256              |
| Consensus     | Proof-of-Work        |
| Data Format   | JSON                 |
| Storage       | `ledger_backup.json` |

---

## 🧠 How It Works

A simplified flow of Nirikshan:

```text
Project / Audit Update
        ↓
   Create Transaction
        ↓
   Add to Blockchain
        ↓
      Mining
        ↓
    Proof-of-Work
        ↓
    SHA-256 Hash
        ↓
      New Block
        ↓
   Persistent Ledger
```

Each block stores information that connects it to the previous block.

```text
Block 1
   ↓
Block 2
   ↓
Block 3
   ↓
Block 4
```

If information inside an earlier block is modified, its hash changes and the chain relationship can be detected.

---

## 📁 Project Structure

```text
Nirikshan/
│
├── main.go
├── index.html
├── go.mod
├── go.sum
├── ledger_backup.json
└── README.md
```

---

# 🚧 Upcoming Updates

Nirikshan is currently being developed further. The following features are planned for future versions.

### 1. 🌐 Multi-Node P2P Blockchain

Move from a single-node implementation toward a real distributed blockchain network.

Planned features:

* Multiple blockchain nodes
* Peer discovery
* Block broadcasting
* Chain synchronization
* Valid-chain selection
* Node health/status monitoring

```text
Government Node
       ↕
Auditor Node
       ↕
Public Observer Node
```

---

### 2. ✍️ Digital Signatures

Introduce cryptographic signatures for authorized actions.

Planned approach:

* Public/private key pairs
* Signed transactions
* Signature verification
* Identity-based transaction authorization

This will allow the system to verify that a transaction was actually created by an authorized participant.

---

### 3. 📦 Dedicated Transaction Layer

Separate transactions from blocks.

Planned transaction structure:

```text
Transaction
├── Transaction ID
├── Project ID
├── Action
├── Amount
├── Timestamp
├── Public Key
└── Digital Signature
```

Multiple transactions can then be collected into a block before mining.

---

### 4. 🧠 Smart-Contract-Like Project Rules

Introduce rules governing valid project state transitions.

Example:

```text
Proposal
   ↓
Approved
   ↓
Funds Released
   ↓
Construction Started
   ↓
Milestone Completed
   ↓
Audited
   ↓
Completed
```

Invalid transitions will be rejected by the system.

---

### 5. 🔍 Blockchain Integrity Verification

Add a dedicated chain validation system.

Planned endpoint:

```text
GET /validate
```

It will verify:

* Block hashes
* Previous-hash relationships
* Proof-of-Work
* Blockchain structure
* Transaction integrity

The dashboard will display whether the chain is currently valid.

---

### 6. 💰 Detailed Financial Transparency

Expand project records to provide a complete financial history.

Example:

```text
Project Budget       ₹50,00,000
Funds Released       ₹30,00,000
Amount Spent         ₹27,50,000
Remaining            ₹20,00,000
```

Every financial event will become a traceable blockchain transaction.

---

### 7. 👥 Role-Based Access Control

Introduce different system roles:

```text
ADMIN
GOVERNMENT
AUDITOR
PUBLIC
```

Each role will have different permissions for creating, updating, auditing, and viewing project information.

JWT-based authentication is planned for API security.

---

### 8. 🧊 Advanced 3D Blockchain Visualization

Expand the existing Three.js dashboard to provide deeper blockchain inspection.

Users will eventually be able to select individual blocks and view:

* Block hash
* Previous hash
* Nonce
* Timestamp
* Transactions
* Project information
* Mining information

---

### 9. 🔎 Public Project Verification

Introduce a unique verification ID for every infrastructure project.

Example:

```text
NIR-2026-00421
```

Anyone will be able to use the ID to inspect:

* Project status
* Budget
* Spending history
* Audit records
* Blockchain history
* Chain integrity

---

# 🎯 Long-Term Vision

The long-term goal of Nirikshan is to evolve from a blockchain demonstration into a distributed infrastructure accountability platform.

```text
             NIRIKSHAN
                 │
       ┌─────────┴─────────┐
       │                   │
   Government            Auditors
       │                   │
       └─────────┬─────────┘
                 │
          Distributed Ledger
                 │
       ┌─────────┴─────────┐
       │                   │
   Project Data       Financial Data
       │                   │
       └─────────┬─────────┘
                 │
           Public Verification
```

The objective is to make infrastructure records **traceable, verifiable, and resistant to unauthorized modification**.

---

## 🚀 Getting Started

### Prerequisites

* Go 1.24+
* Git

### Installation

```bash
git clone https://github.com/atharvtamboli/nirikshan_ai.git
cd nirikshan_ai
go mod tidy
```

### Run

```bash
go run main.go
```

Then open the frontend through the application's configured interface.

---

## 📌 Current Development Status

**Current Version:** Core Blockchain Prototype

### Implemented

* Blockchain data structure
* Block creation
* SHA-256 hashing
* Proof-of-Work mining
* Project recording
* Audit/update records
* JSON ledger persistence
* Go/Gin backend
* Three.js visualization
* Web dashboard

### In Development / Planned

* Multi-node P2P network
* Digital signatures
* Transaction architecture
* Chain validation
* Financial tracking
* Role-based authentication
* Smart-contract-like validation
* Public project verification
* Advanced blockchain visualization

---

## 📄 License

This project is currently maintained as an independent development project by **Atharv Tamboli**.
