# Tasks: OpenSearchProtocol (OSP)

## 1. Executive Summary
This document provides a granular, step-by-step task list for the implementation of the OpenSearchProtocol (OSP). The tasks are categorized by system module and project phase, following Go best practices and a distributed architecture.

---

## 2. 🏗 High-Level Infrastructure

### 2.1 Project Initialization
- [x] **Git Setup:** Initialize Git repository and add a comprehensive `.gitignore`.
- [x] **Go Setup:** Run `go mod init github.com/zulfikawr/go-search`.
- [x] **Directory Structure:** Create standard Go layout.
- [x] **Dependency Management:** Add initial dependencies (gRPC, Protobuf, Fiber).

### 2.2 Logging & Observability
- [x] **Logger Implementation:** JSON-structured logging with `log/slog`.
- [x] **Trace Middleware:** (Context-aware logging implemented).
- [x] **Metrics Setup:** Prometheus metrics for latency, worker count, and errors.

---

## 3. 📡 Protocol Layer (gRPC)

### 3.1 Protobuf Definition (`proto/search.proto`)
- [x] **Bi-directional Stream:** Worker status and Master commands.
- [x] **Metadata Support:** Scrape latency, worker region, and source engine tagging.
- [x] **Code Generation:** Automated via `scripts/gen-proto.sh`.

### 3.2 Security (mTLS)
- [x] **Certificate Generator:** Root CA and node certificate logic in `pkg/crypto`.
- [x] **TLS Config:** Secure gRPC transport in `pkg/protocol/tls.go`.

---

## 4. 🧠 Master Node Implementation

### 4.1 Worker Registry (`internal/master/registry.go`)
- [x] **Dynamic Registry:** Thread-safe worker tracking with Region support.
- [x] **Self-Healing:** Automatic pruning of inactive workers.

### 4.2 Scheduler & Dispatcher
- [x] **Engine-Aware Scheduler:** Load balances across workers supporting specific engines.
- [x] **Async Dispatcher:** Manages task correlation and timeouts.

### 4.3 Result Aggregator & Metadata
- [x] **URL Normalization:** Deduplicates results by stripping tracking parameters.
- [x] **Metadata Injection:** Supports structured OSP signals and performance data.

### 4.4 HTTP API Layer
- [x] **Brave API Parity:** `/web/search` endpoint with 1:1 schema compatibility.
- [x] **Opt-in Verbosity:** `?metadata=true` toggle for OSP extended metrics.

---

## 5. 🦾 Worker Node Implementation

### 5.1 Connectivity & Registration
- [x] **Persistent gRPC:** Auto-registration and heartbeat streaming.
- [x] **Regional Identity:** Workers report their geographic region.

### 5.2 Scraper Pool (`internal/worker/scraper/`)
- [x] **Engine Interface:** Unified scraping contract.
- [x] **Scrapers:** Google, Brave, and DuckDuckGo (HTML) operational.
- [x] **URL Cleaning:** Extracts clean destination URLs from redirects.

### 5.3 Fingerprinting & Stealth
- [x] **TLS Spoofing (JA3):** Chrome v120+ fingerprinting using uTLS.
- [x] **Browser Emulation:** Full header suite and User-Agent rotation.

---

## 6. 🧪 Testing & Validation
- [x] **Integration Testing:** `scripts/run-demo.sh` verified live cluster and data flow.
- [x] **Metadata Verification:** Verified `?metadata=true` logic.

---

## 7. 🚀 DevOps & Deployment
- [ ] **Dockerization:** Container images for Master and Worker.
- [ ] **Docker Compose:** One-liner cluster setup.

---

## 8. 📝 Documentation
- [ ] **README:** Final project documentation.
