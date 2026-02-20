# GOSP Architecture 🦞

GOSP follows a **Distributed Orchestration** pattern. It separates the "intent" (the search query) from the "acquisition" (the web scrape) to ensure maximum resilience and bypass IP-based blocking.

## The Two-Tier Model

### 1. The Master Node (The Brain)
The Master acts as the public-facing gateway and the network orchestrator.
- **HTTP API:** Exposes a Brave Search API compatible endpoint.
- **Worker Registry:** Tracks every connected worker, their health, and their geographic region.
- **Task Dispatcher:** Receives a search request and selects the best worker(s) to perform the task.
- **Result Aggregator:** Merges data from multiple workers, deduplicates URLs, and ranks the results before serving the final JSON.

### 2. The Worker Node (The Hands)
Workers are lightweight scraping services meant to be deployed on diverse network environments.
- **Stealth Engine:** Uses JA3 TLS fingerprinting to mimic a real Chrome/Safari browser.
- **Scraper Pool:** Implements a variety of engine-specific scrapers (Google, Brave, DuckDuckGo).
- **gRPC Stream:** Maintains a persistent, bi-directional connection to the Master to bypass firewalls (NAT) without port forwarding.

---

## Data Flow Lifecycle

1.  **Request:** A user sends a `GET /web/search?q=test` to the Master.
2.  **Scheduling:** Master looks at the healthy workers and picks an available one that supports the requested engine.
3.  **Dispatch:** Master sends a `SearchTask` gRPC message to the Worker.
4.  **Acquisition:** Worker performs a raw HTTP scrape of the search engine, cleans the data, and extracts the URLs.
5.  **Return:** Worker sends a `SearchResponse` back to the Master.
6.  **Aggregation:** Master cleans the results, injects metadata (if requested), and returns the JSON to the user.

---

## Why Distributed?

Standard scrapers running on a single VPS (Azure, AWS, GCP) are blocked by Google/Brave within minutes. GOSP solves this by allowing you to run the **Master** on a VPS but the **Workers** on residential IPs (home PCs, Raspberry Pis, Mobile phones). 

Search engines see the traffic coming from "real users" at home, making the GOSP network virtually unblockable.
