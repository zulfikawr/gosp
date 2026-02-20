# OpenClaw + GOSP Integration Guide 🦞

This guide explains how to use GOSP as a zero-cost, drop-in replacement for the Brave Search API within OpenClaw.

## How it Works
AI agents like OpenClaw have built-in support for Brave Search but require a paid API key. GOSP "hijacks" these requests by:
1.  Redirecting `api.search.brave.com` to your local GOSP Master via `/etc/hosts`.
2.  Running the GOSP Master on Port 443 with a self-signed TLS certificate.
3.  Implementing a 1:1 compatible Brave Search API response schema.

## Setup Instructions

### 1. System Redirect
Map the official Brave API domain to your local machine:
```bash
echo "127.0.0.1 api.search.brave.com" | sudo tee -a /etc/hosts
```

### 2. Prepare TLS Certificates
GOSP Master must serve HTTPS on port 443. Generate a self-signed certificate in the GOSP root directory:
```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 3650 -nodes -subj "/CN=api.search.brave.com"
```

### 3. Start GOSP Cluster
**Master (Brain):**
Run as root to bind to port 443. It will automatically detect `cert.pem` and `key.pem` to enable HTTPS.
```bash
sudo ./gosp master run main --no-daemon
```

**Worker (Scraper):**
```bash
./gosp worker run local-01 --no-daemon
```

### 4. Configure OpenClaw
Edit your `~/.openclaw/openclaw.json` to enable the Brave provider. The API key can be any string.

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

## Verification
You can verify the integration by running a curl command. It should return results from your local workers:
```bash
curl -k "https://api.search.brave.com/res/v1/web/search?q=openclaw"
```
