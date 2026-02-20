# GOSP Security & Protocol 🦞

Security is a primary pillar of GOSP. As a distributed system, it must protect the Master from malicious workers and ensure the search data remains integral.

## Authentication: The Join Token

GOSP uses a wordlist-based Join Token system (inspired by PAKE and Tailscale) to secure the cluster.

1.  **Generation:** When a Master is created, it generates a persistent token (e.g., `ghost-mountain-zebra`).
2.  **Handshake:** Workers must provide this token during the `Connect` phase.
3.  **gRPC Interceptor:** The Master implements a security middleware that checks every incoming gRPC call. If the `authorization` metadata is missing or incorrect, the Master instantly drops the connection with an `Unauthenticated` status.

## Transport: Secure gRPC

- **Bi-directional Streaming:** GOSP uses long-lived gRPC streams. This allows workers to initiate the connection (outbound), enabling them to work behind residential NAT/Firewalls without port forwarding.
- **TLS:** All traffic is encrypted using TLS 1.3. GOSP supports both insecure development modes and full mTLS (Mutual TLS) for production clusters.

## Advanced Stealth Layer

Every Worker node is equipped with a **Stealth Engine** (`pkg/stealth`) to bypass automated bot detection:

### 1. JA3 Fingerprinting
Standard Go `net/http` clients are easily detected because their TLS handshakes follow a predictable pattern. GOSP uses **uTLS** to spoof the **JA3 fingerprint** of modern browsers like **Chrome v120**. 

### 2. Header Ordering
Search engines like Google look at the order of HTTP headers. GOSP ensures that headers are sent in the exact sequence expected by a real Chrome browser.

### 3. User-Agent Rotation
Workers rotate through a pool of high-trust User-Agents corresponding to the spoofed TLS fingerprints.
