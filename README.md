# Nirikshan | Autonomous Infrastructure Ledger 🏗️

**Nirikshan** is a Go-based blockchain application designed to ensure transparency and immutability in public infrastructure projects. By recording every project proposal and financial audit on a decentralized ledger, it prevents unauthorized tampering with government records and ensures high-level accountability.



## 🌟 Key Features
* **Proof-of-Work Consensus:** Implements a mining algorithm with adjustable difficulty to secure the chain.
* **Immutable Audit Trails:** Every project update (spending, status changes) is recorded as a new block, preserving the entire history of the project.
* **3D Network Visualization:** A futuristic dashboard built with Three.js that renders the blockchain topology in real-time.
* **Persistence Layer:** Automatically backups the ledger to `ledger_backup.json` to prevent data loss across sessions.
* **Glassmorphism UI:** A high-end, responsive frontend for committing projects and performing audits.

## 🛠️ Technical Stack
* **Backend:** Go (Golang) using the Gin Web Framework.
* **Frontend:** HTML5, Tailwind CSS, and Three.js.
* **Cryptography:** SHA-256 hashing for block validation and AES encryption for data security.
* **Data Format:** JSON-based block structure for seamless API integration.

## 🚀 Getting Started

### 1. Prerequisites
* Go 1.24+ installed on your machine.

### 2. Installation
```bash
git clone [https://github.com/atharvtamboli/Nirikshan.git](https://github.com/atharvtamboli/Nirikshan.git)
cd Nirikshan
go mod tidy
