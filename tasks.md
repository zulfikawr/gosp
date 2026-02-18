# Tasks: OpenSearchProtocol (OSP)

## 1. Executive Summary
This document provides a granular, step-by-step task list for the implementation of the OpenSearchProtocol (OSP). The tasks are categorized by system module and project phase, following Go best practices and a distributed architecture.

---

## 2. 🏗 High-Level Infrastructure

### 2.1 Project Initialization
- [ ] **Git Setup:** Initialize Git repository and add a comprehensive `.gitignore`.
- [ ] **Go Setup:** Run `go mod init github.com/zulfikawr/go-search`.
- [ ] **Directory Structure:** Create the following:
    - `cmd/master/main.go` - The entry point for the Master node.
    - `cmd/worker/main.go` - The entry point for the Worker node.
    - `internal/master/` - All private Master logic (handlers, registry, scheduler).
    - `internal/worker/` - All private Worker logic (client, scraper pool).
    - `pkg/models/` - Shared public structs (SearchRequest, SearchResponse).
    - `pkg/protocol/` - gRPC logic and TLS management.
    - `proto/` - Protobuf source files.
    - `scripts/` - Shell scripts for build/proto generation.
- [ ] **Dependency Management:** Add initial dependencies (`google.golang.org/grpc`, `google.golang.org/protobuf`, `github.com/gofiber/fiber/v2`).

### 2.2 Logging & Observability
- [ ] **Logger Implementation:** Wrap `log/slog` in a custom `pkg/logger` that supports JSON formatting and dynamic log levels.
- [ ] **Trace Middleware:** Implement a simple trace ID middleware for Fiber and gRPC to link requests across nodes.
- [ ] **Metrics Setup:** Integrate `prometheus/client_golang` for P95 latency and worker count tracking.

---

## 3. 📡 Protocol Layer (gRPC)

### 3.1 Protobuf Definition (`proto/search.proto`)
- [ ] **Message `SearchRequest`:** Include `query`, `engine`, `count`, `offset`, and metadata.
- [ ] **Message `SearchResponse`:** Include `results` (repeated list of titles/URLs/snippets) and `meta` (latency, source worker).
- [ ] **Service `SearchService`:**
    - `rpc Register(RegisterRequest) returns (RegisterResponse)`
    - `rpc Heartbeat(stream HeartbeatRequest) returns (stream HeartbeatResponse)`
    - `rpc Fetch(SearchRequest) returns (SearchResponse)`
- [ ] **Code Generation:** Create `scripts/gen-proto.sh` and generate Go files.

### 3.2 Security (mTLS)
- [ ] **Certificate Generator:** Implement a Go utility to generate a self-signed CA and worker/master certificates for testing.
- [ ] **TLS Config:** Create `pkg/protocol/tls.go` to handle loading certs into gRPC credentials.

---

## 4. 🧠 Master Node Implementation

### 4.1 Worker Registry (`internal/master/registry.go`)
- [ ] **Data Structure:** Use a thread-safe `map[string]*WorkerNode`.
- [ ] **Registration:** Implement the logic to add/update workers upon connection.
- [ ] **Health Check:** Implement a background goroutine to remove workers that haven't sent a heartbeat for > 60 seconds.

### 4.2 Scheduler & Dispatcher
- [ ] **Round-Robin Scheduler:** Basic load balancing for search tasks.
- [ ] **Task Dispatcher:** The logic to select an available worker, send the gRPC `Fetch` call, and handle the response.
- [ ] **Timeout Handling:** Wrap every worker call in a `context.WithTimeout`.

### 4.3 Result Aggregator
- [ ] **Deduplication:** Implement URL-based deduplication using a `map[string]bool`.
- [ ] **Ranking v1:** A simple heuristic-based ranker to order results from different engines.
- [ ] **Brave API Mapping:** The final step to convert internal results into the official Brave Search API JSON format.

### 4.4 HTTP API Layer
- [ ] **Fiber Server:** Set up the main Fiber app in `cmd/master/main.go`.
- [ ] **Endpoint `/web/search`:** The primary public API endpoint.
- [ ] **Request Validation:** Use `go-playground/validator` for incoming search parameters.

---

## 5. 🦾 Worker Node Implementation

### 5.1 Connectivity & Registration
- [ ] **gRPC Client:** Implement the persistent connection logic in `internal/worker/client.go`.
- [ ] **Auto-Registration:** Upon startup, the worker must send its metadata to the Master.
- [ ] **Heartbeat Stream:** Open a bidirectional gRPC stream for real-time status updates.

### 5.2 Scraper Pool (`internal/worker/scraper/`)
- [ ] **Engine Interface:** Define `Scraper` with a `Search(query) ([]Result, error)` method.
- [ ] **Google Engine:** Raw HTTP scraper using `net/http` and `PuerkitoBio/goquery`.
- [ ] **Brave Engine:** Scraper for the Brave web search results.
- [ ] **Bing Engine:** Scraper for Microsoft Bing.
- [ ] **Proxy Support:** (Optional) Logic to route requests through a local SOCKS5/HTTP proxy.

### 5.3 Fingerprinting & Stealth
- [ ] **User-Agent Rotation:** Implement a random UA picker from a curated list.
- [ ] **TLS Fingerprinting:** (Optional) Use `utls` to mimic Chrome/Firefox TLS handshakes.

---

## 6. 🧪 Testing & Validation

### 6.1 Unit Testing
- [ ] **Scraper Tests:** Test each scraper against local HTML snapshots (`internal/worker/scraper/google_test.go`).
- [ ] **Aggregator Tests:** Verify that results are correctly merged and deduplicated.
- [ ] **Registry Tests:** Ensure workers are correctly added and pruned.

### 6.2 Integration Testing
- [ ] **End-to-End Flow:** A script that spins up a Master, 2 Workers, and runs a real query from a CLI tool.
- [ ] **Failure Scenario:** Kill a worker and verify the Master retries on the second worker.

### 6.3 Performance Testing
- [ ] **Load Test:** Use `k6` or `vegeta` to measure the Master node's performance under load.
- [ ] **Latency Profile:** Measure P50, P90, and P99 response times.

---

## 7. 🚀 DevOps & Deployment

### 7.1 Dockerization
- [ ] **Master Dockerfile:** Standard multi-stage build.
- [ ] **Worker Dockerfile:** Optimized for ARMv7 and ARM64 (Raspberry Pi).
- [ ] **Compose File:** A `docker-compose.yaml` for a local 3-node cluster.

### 7.2 CI/CD (GitHub Actions)
- [ ] **Go Lint:** Run `golangci-lint` on every commit.
- [ ] **Build & Test:** Build all binaries and run all tests.
- [ ] **Image Registry:** Push successful builds to GHCR (`ghcr.io/zulfikawr/go-search`).

---

## 8. 📝 Documentation
- [ ] **README:** Clear overview, architecture, and "Getting Started."
- [ ] **API Spec:** OpenAPI/Swagger documentation.
- [ ] **Contributor Guide:** How to add a new engine or fix a bug.
- [ ] **Security Policy:** How to report vulnerabilities.
