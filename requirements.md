# OpenSearchProtocol (OSP): Technical Requirements Specification

## 1. Project Overview
OpenSearchProtocol (OSP) is a distributed, high-performance, and decentralized search scraping protocol written in Go. Its goal is to provide a free, open-source alternative to the Brave Search API by leveraging a network of volunteer-operated worker nodes running on residential IP addresses.

### 1.1 Core Mission
- **Decentralization:** No single point of failure for search data acquisition.
- **Sovereignty:** Users can run their own Master and Worker nodes.
- **Interoperability:** Drop-in replacement for existing Brave Search API clients.
- **Scalability:** Handle thousands of concurrent workers and requests.

---

## 2. Functional Requirements

### 2.1 Master Node (API & Orchestrator)
The Master node is the central API gateway. It receives HTTP requests, breaks them down into sub-tasks, distributes them to workers, and aggregates the results.

#### 2.1.1 API Compatibility
- **Endpoint Parity:** Must implement `/web/search` matching the Brave Search API.
- **Response Schema:** JSON responses must be 1:1 compatible with the Brave Search API's `SearchResponse` object.
- **Parameters:** Support `q` (query), `country`, `search_lang`, `ui_lang`, `count`, `offset`, `safesearch`, `freshness`, and `text_decorations`.

#### 2.1.2 Task Orchestration
- **Worker Discovery:** Master must track all active and healthy workers.
- **Task Scheduling:** Implement a multi-strategy scheduler (Round-Robin, Latency-Aware, or Geographic).
- **Redundancy:** Every query must be sent to at least $N$ workers (configurable) to ensure result reliability.
- **Timeout Management:** Strict context-based timeouts for all worker interactions.

#### 2.1.3 Data Processing
- **Deduplication:** Merging results from multiple workers and removing duplicates based on URL.
- **Ranking:** A basic internal ranker to order the aggregated results.
- **Caching:** An optional Redis-backed caching layer for high-frequency queries.

### 2.2 Worker Node (Scraper & Proxy)
The Worker node is a lightweight binary that performs the actual search engine interaction.

#### 2.2.1 Scraper Engines
- **Initial Support:** Google, Bing, Brave, and DuckDuckGo.
- **Pluggable Architecture:** Ability to add new "providers" by implementing a Go interface.
- **Scraping Strategies:**
    - **Raw HTTP:** High-speed, low-resource scraping with custom TLS fingerprints.
    - **Headless (Optional):** Using `chromedp` for engines requiring JavaScript execution.

#### 2.2.2 Connectivity
- **Persistency:** Workers must maintain a persistent gRPC stream to the Master.
- **Handshake:** Automated registration with the Master using a shared secret or public key.
- **Heartbeats:** Regular status updates (CPU, memory, IP status) to the Master.

---

## 3. Technical Protocol (The OSP Standard)

### 3.1 gRPC Definition
All communication between Master and Worker must happen over gRPC using Protocol Buffers (Protobuf).

#### 3.1.1 SearchTask Message
- `string task_id`: Unique UUID for the task.
- `string query`: The search string.
- `enum Engine`: Target engine (Google, Bing, etc.).
- `map<string, string> params`: Additional search parameters.
- `google.protobuf.Timestamp deadline`: Time when the task expires.

#### 3.1.2 SearchResult Message
- `string task_id`: Reference to the original task.
- `repeated ResultItem results`: List of scraped search results.
- `string raw_html_sample`: (Optional) Sample for verification.
- `uint32 latency_ms`: Time taken by the worker to perform the scrape.
- `enum ErrorCode`: Success, RateLimited, Captcha, Timeout, or ProviderDown.

### 3.2 Security Model
- **mTLS:** All gRPC traffic must be encrypted via mutual TLS (mTLS).
- **Authentication:** Master must verify Worker's public key before accepting results.
- **Sandboxing:** Workers must never execute arbitrary code from the Master.
- **Rate Limiting:** Master must limit how many tasks a single worker can receive per second to prevent worker abuse.

---

## 4. Non-Functional Requirements

### 4.1 Performance Targets
- **Master API Latency:** 95th percentile (P95) response time < 1,200ms.
- **Worker Scrape Latency:** < 800ms per engine.
- **Master Throughput:** Support 500+ requests per second on standard VPS hardware.
- **Worker Footprint:** < 50MB RAM and < 5% CPU on idle.

### 4.2 Reliability & Fault Tolerance
- **Circuit Breaking:** Master must automatically "quarantine" a worker if it returns 3 consecutive errors.
- **Retry Logic:** Automatic retry of failed tasks on different workers up to 3 times.
- **State Persistence:** Master should survive a restart without losing the worker registry.

### 4.3 Observability
- **Structured Logging:** All components must use `log/slog`.
- **Metrics:** Prometheus-compatible metrics for:
    - Active worker count.
    - Success/Failure rate per engine.
    - Latency per worker.
    - Total queries served.
- **Tracing:** (Future) OpenTelemetry tracing for the entire request lifecycle.

---

## 5. Developer Experience (DX)

### 5.1 CLI Tooling
- `osp-master`: The main server binary.
- `osp-worker`: The worker binary.
- `osp-cli`: A management tool to inspect the cluster, ban workers, and run manual test queries.

### 5.2 Documentation
- **API Spec:** OpenAPI/Swagger documentation for the HTTP API.
- **Deployment Guide:** Docker Compose and Systemd configurations.
- **Contributor Guide:** How to add a new scraper engine in Go.

---

## 6. Constraints & Compliance

### 6.1 Legal & Privacy
- **User Privacy:** Master must strip identifying information (IP, User-Agent) from requests before sending them to Workers.
- **ToS Compliance:** The protocol must prioritize "ethical scraping" (low frequency, respecting robots.txt where possible).
- **Transparency:** The project must be licensed under a permissive open-source license (MIT/Apache 2.0).

### 6.2 Technical Constraints
- **Language:** Must be written in pure Go (no CGO preferred for easy cross-compilation).
- **Dependencies:** Minimize third-party dependencies to keep the binary size small.
- **Cross-Platform:** Workers must run on Linux, macOS, and Windows (specifically Raspberry Pi/ARM support).

---

## 7. Future Considerations
- **Proof of Work (PoW):** Prevent worker spamming by requiring a small PoW for registration.
- **Economic Incentives:** Tokenized rewards for worker uptime (optional module).
- **Zero-Knowledge Proofs:** Validating results without the Master knowing which worker did what.
- **Libp2p Integration:** Full P2P discovery to remove the need for a central Master IP.

---

## 8. Detailed API Specification (Brave Search API v1 Parity)

The Master node must implement the following HTTP endpoints to ensure drop-in compatibility.

### 8.1 GET `/web/search`

#### Request Parameters
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `q` | string | Yes | The search query string. |
| `country` | string | No | 2-letter country code (ISO 3166-1 alpha-2). |
| `search_lang` | string | No | Search language code (ISO 639-1). |
| `ui_lang` | string | No | UI language code (ISO 639-1). |
| `count` | integer | No | Number of results (max 20). |
| `offset` | integer | No | Pagination offset. |
| `safesearch` | string | No | `off`, `moderate`, `strict`. |
| `freshness` | string | No | `pd` (past day), `pw` (past week), `pm`, `py`. |
| `text_decorations` | boolean | No | Whether to include <b> tags in snippets. |

#### Response Schema (Brave Parity)
```json
{
  "type": "search",
  "query": {
    "original": "string",
    "altered": "string",
    "canonical": "string"
  },
  "web": {
    "type": "search_results",
    "results": [
      {
        "title": "string",
        "url": "string",
        "description": "string",
        "age": "string",
        "language": "string",
        "thumbnail": {
          "src": "string"
        },
        "profile": {
          "name": "string",
          "url": "string",
          "img": "string"
        }
      }
    ]
  },
  "news": {
    "results": []
  },
  "videos": {
    "results": []
  },
  "images": {
    "results": []
  }
}
```

---

## 9. Standardized OSP Error Codes

All errors in the OSP ecosystem (Master-Worker and Master-Client) must map to these codes.

| Code | Name | Description |
| :--- | :--- | :--- |
| `1000` | `SUCCESS` | Task completed successfully. |
| `2001` | `INVALID_QUERY` | Search query is empty or malformed. |
| `3001` | `WORKER_TIMEOUT` | Worker failed to respond within the deadline. |
| `3002` | `WORKER_DISCONNECT` | Worker lost connection during task execution. |
| `4001` | `RATE_LIMITED` | Upstream engine (Google/Brave) rate-limited the worker. |
| `4002` | `CAPTCHA_DETECTED` | Worker encountered a CAPTCHA. |
| `4003` | `PROVIDER_DOWN` | Upstream engine is unresponsive. |
| `5001` | `INTERNAL_MASTER_ERROR` | Unexpected error in the Master node logic. |

---

## 10. Security & Authentication (mTLS Deep Dive)

To reach "Industry Standard," OSP does not use simple API keys for Master-Worker communication. It uses Mutual TLS (mTLS).

### 10.1 Handshake Sequence
1. **Worker Start:** Generates an Ed25519 key pair.
2. **Registration:** Worker sends its Public Key to the Master via an out-of-band channel or a "Bootstrap Key."
3. **mTLS Setup:** Master and Worker establish a TLS 1.3 connection where both sides provide certificates signed by the OSP Cluster CA.
4. **Validation:** Master verifies that the certificate's Common Name (CN) matches the registered Worker ID.

### 10.2 Data Integrity
- **Result Signing:** Every `SearchResult` protobuf is digitally signed by the Worker's private key.
- **Timestamping:** Results include a high-precision timestamp to prevent "Replay Attacks."

---

## 11. Performance Budget (Latency Breakdown)

Total Budget: **1,200ms** (P95)

| Step | Max Time | Responsibility |
| :--- | :--- | :--- |
| Request Parsing | 10ms | Master |
| Worker Selection | 5ms | Master |
| gRPC Round Trip | 50ms | Network |
| Upstream Scrape | 800ms | Worker |
| Result Parsing | 100ms | Worker |
| Aggregation/Ranking | 100ms | Master |
| JSON Serialization | 20ms | Master |
| **Total** | **1,085ms** | |

---

## 12. Networking Requirements (Libp2p)

In Phase 4, OSP will migrate to `libp2p` to support residential workers behind NAT.

- **Transport:** TCP/QUIC (UDP).
- **Discovery:** Distributed Hash Table (DHT) or MDNS (Local).
- **Relay:** Use AutoRelay (Circuit Relay v2) to allow workers with closed ports to be reachable via a "Relay Node."
- **Encryption:** Noise Protocol (default in libp2p).

---

## 13. Logging & Monitoring (Slog Specification)

Master and Worker must emit logs in JSON format using `log/slog`.

### 13.1 Required Fields
- `time`: RFC3339 timestamp.
- `level`: `INFO`, `WARN`, `ERROR`, `DEBUG`.
- `msg`: Human-readable message.
- `component`: `master`, `worker`, `scraper`, `api`.
- `trace_id`: UUID linking HTTP request to gRPC tasks.

### 13.2 Example Log Entry
```json
{
  "time": "2026-02-18T21:45:00Z",
  "level": "INFO",
  "msg": "task_assigned",
  "worker_id": "worker-01-us",
  "task_id": "abc-123",
  "query": "golang gRPC",
  "engine": "google"
}
```
