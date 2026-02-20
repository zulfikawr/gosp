# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-02-20

### Added
- **Distributed Architecture:** Initial release of the Master (Orchestrator) and Worker (Scraper) model.
- **Protocol:** High-performance bi-directional gRPC streaming for Master-Worker communication.
- **API Compatibility:** 1:1 JSON response parity with Brave Search API v1.
- **Stealth Engine:** JA3/uTLS fingerprinting (Chrome v120) and header spoofing to bypass bot detection.
- **Security:** Secure, wordlist-based Join Token authentication and gRPC interceptors.
- **Service Manager CLI:** Unified `gosp` binary with profile management (`master create`, `worker run`, etc.).
- **Background Execution:** Integrated daemon mode support for both Master and Worker nodes.
- **Transparency:** Opt-in metadata flag (`?metadata=true`) for worker region and performance breakdown.
- **Scraper Pool:** Initial support for DuckDuckGo (HTML), Google (Mobile), and Brave Search engines.
- **Developer Tooling:** Standardized `Makefile` and comprehensive `docs/` suite.

---
[0.1.0]: https://github.com/zulfikawr/gosp/releases/tag/v0.1.0
