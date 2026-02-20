# GOSP Scraping Engines 🦞

GOSP Workers maintain a pool of specialized search engine scrapers. Unlike "Yes-Agents," GOSP acknowledges the technical difficulty of web scraping and implements multiple fallback strategies.

## Supported Engines

### 1. DuckDuckGo (The Reliable Fallback)
DuckDuckGo is GOSP's default and most reliable engine.
- **Strategy:** Uses the HTML-only endpoint.
- **Reliability:** High. It is very lenient towards datacenter IPs and raw HTTP scrapers.
- **Cleanup:** GOSP automatically cleans the DuckDuckGo internal redirects (`uddg=`) to return clean destination URLs.

### 2. Google (The Residential Target)
Google is the primary target for GOSP but is highly protected.
- **Strategy:** Mobile Safari Emulation.
- **The VPS Wall:** Google strictly blacklists data-center IP ranges. Running a GOSP Worker on a VPS for Google search will often return empty results or bot-detection walls.
- **The Solution:** To get real Google results, connect a GOSP Worker from a **Residential (Home) IP**.

### 3. Brave Search (The Meta Target)
Brave Search provides high-quality data but uses modern bot protection.
- **Strategy:** Clean Web view parsing.
- **Stealth:** Requires JA3 fingerprinting to bypass initial handshake checks.

## Pluggable Interface

Every scraper implements the `Engine` interface in `internal/worker/scraper/interface.go`:

```go
type Engine interface {
    Search(query string, count int32, offset int32) ([]*protocol.ResultItem, error)
    ID() protocol.Engine
}
```

New engines can be added by implementing this interface and registering them in the Worker client.
