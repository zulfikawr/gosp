# GOSP (Go OpenSearchProtocol) 🦞

**A secure, distributed, and unblockable search protocol designed for the next generation of web scraping.**

GOSP provides a free, open-source alternative to centralized search APIs (like Brave or Google) by leveraging a network of distributed Worker nodes. It is designed to bypass data-center IP blacklists by performing search acquisition from residential IP addresses while providing a 1:1 compatible Brave Search API interface.

---

## 🚀 Key Features

*   **Distributed Architecture:** Decoupled Master (Orchestrator) and Worker (Scraper) model.
*   **Brave Search Parity:** 1:1 JSON response schema compatibility—a true drop-in replacement.
*   **Advanced Stealth:** TLS Fingerprinting (JA3/uTLS) and browser-grade header spoofing to bypass bot detection.
*   **Secure by Design:** Handshake secured by cryptographically verified wordlist Join Tokens and gRPC interceptors.
*   **Service Manager CLI:** daemon-by-default background execution with profile-based management.
*   **Transparent Metadata:** Opt-in detailed metrics including worker region, engine source, and latency breakdowns.

---

## 🛠 Quick Start

### 1. Installation
Build the unified GOSP binary using the included Makefile:
```bash
git clone https://github.com/zulfikawr/gosp
cd gosp
make build
```

### 2. Initialize the Brain (Master)
Create a profile for your Master node and start it in the background:
```bash
./bin/gosp master create
./bin/gosp master run
```

### 3. Initialize the Hands (Worker)
Connect a local scraper to your Master using the Join Token generated during master creation:
```bash
./bin/gosp worker create
./bin/gosp worker run
```

### 4. Search the Cluster
Perform high-speed searches directly from your terminal:
```bash
./bin/gosp search --query "distributed systems"
```

---

## 📁 Documentation

Dive deeper into the GOSP ecosystem:

*   [**Architecture Overview**](docs/architecture.md): How the distributed protocol works.
*   [**CLI Reference**](docs/commands.md): Detailed usage of `master`, `worker`, and `search`.
*   [**Security & Protocol**](docs/security.md): mTLS and Join Token technical details.
*   [**API Specification**](docs/api.md): Brave Search API compatibility and Metadata flags.
*   [**Scraping Engines**](docs/scraping.md): How Google, Brave, and DuckDuckGo are handled.
*   [**OpenClaw Integration**](docs/openclaw.md): Use GOSP as a free Brave Search API replacement.

---

## 🦞 GOSP as a Brave Search API Alternative (OpenClaw Setup)

You can use GOSP to power AI agents (like OpenClaw) without paying for API keys by "hijacking" the Brave Search API endpoint.

### 1. Configure System Redirect
Since AI agents hardcode the Brave API URL, you must redirect that traffic to your local GOSP Master.
Add this line to your `/etc/hosts` file:
```text
127.0.0.1 api.search.brave.com
```

### 2. Start GOSP Master on Port 443
The Brave API expects HTTPS (Port 443). Start the Master with root privileges:
```bash
sudo ./gosp master run --addr 0.0.0.0:443
```

### 3. Start GOSP Worker
```bash
./gosp worker run --master localhost:443 --id local-ubuntu
```

### 4. Configure OpenClaw
GOSP implements the Brave-compatible schema. Update your `~/.openclaw/openclaw.json`:
```json
{
  "tools": {
    "web": {
      "search": {
        "enabled": true,
        "provider": "brave",
        "apiKey": "gosp_free_tier"
      }
    }
  }
}
```

**Why this works:** OpenClaw will try to talk to `api.search.brave.com:443`. Your system will redirect that call to your GOSP Master running locally on port 443. GOSP then processes the search for free using its distributed workers.

---

## ⚖️ License
GOSP is licensed under the MIT License. See [LICENSE](LICENSE) for details.
