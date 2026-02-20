# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.1] - 2026-02-21

### Fixed
- **API Stability:** Fixed a crash in `/cluster/status` caused by JSON serialization of Go channels in the Worker struct.
- **Token Security:** Expanded the `pkg/tokens` wordlist from 32 to 96 words to significantly reduce the probability of Join Token collisions.

### Added
- **Test Suite:** Implemented a comprehensive test suite for all critical packages:
    - **E2E API Tests:** Verified Brave-compatible search flow and cluster status.
    - **Stealth Transport:** Unit tests for anti-detection HTTP client.
    - **Core Packages:** Achieved >80% coverage for `config`, `logger`, `tokens`, and `pid`.

## [0.2.0] - 2026-02-21

### Added
- **AI Agent Integration:** Added `docs/openclaw.md` for seamless, free Brave Search API "hijacking" with OpenClaw.
- **TLS Support:** Integrated `ListenTLS` in HTTPServer to allow Master nodes to serve over Port 443 with custom certificates.
- **Brave API Compatibility:** Added explicit `/res/v1/web/search` route for 1:1 drop-in compatibility with the official Brave Search API endpoint.
- **Protocol Interception:** Added support for local domain redirection (`api.search.brave.com`) to enable zero-cost search for AI agents.

### Fixed
- **Daemon Spawning:** Refactored background process logic to use `os.Executable()` for more reliable binary resolution.
- **File Handling:** Added proper error checking for log file creation and initialization during daemon startup.
- **Security:** Updated `.gitignore` to prevent accidental leakage of local TLS `.pem` certificates.

### Changed
- **CLI Flags:** Enhanced `master run` to automatically detect and use `cert.pem`/`key.pem` if present in the working directory.

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
[0.2.1]: https://github.com/zulfikawr/gosp/releases/tag/v0.2.1
[0.2.0]: https://github.com/zulfikawr/gosp/releases/tag/v0.2.0
[0.1.0]: https://github.com/zulfikawr/gosp/releases/tag/v0.1.0
