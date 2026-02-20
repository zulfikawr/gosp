# GOSP CLI Reference 🦞

GOSP is managed through a unified binary that follows a "Service Manager" model. By default, it runs in the background (daemon) using profiles stored in `~/.gosp/`.

## Root Commands

### `gosp master`
Manage orchestrator nodes.
- **`create`**: Interactively create a Master profile. Generates ports and Join Tokens.
- **`run`**: Start a Master in the background. Use `--no-daemon` for foreground.
- **`stop`**: Gracefully terminate a running Master.
- **`status`**: Show Master uptime and a list of all connected Workers.
- **`list`**: List all configured Master profiles.
- **`delete <name>`**: Remove a profile.

### `gosp worker`
Manage scraping nodes.
- **`create`**: Interactively connect to a Master. Requires a valid Join Token.
- **`run`**: Start the scraper in the background.
- **`stop`**: Gracefully disconnect and stop the worker.
- **`status`**: Show if the local worker is connected to its Master.
- **`list`**: List local worker profiles.

### `gosp search`
Query the GOSP cluster.
- **`-q, --query <string>`**: The search term (Required).
- **`-e, --engine <google|duckduckgo>`**: Select search engine (Default: duckduckgo).
- **`-c, --count <int>`**: Number of results (Default: 10).
- **`-m, --metadata`**: Enable OSP extended metadata (region, worker ID).
- **`-f, --format <table|json>`**: Output format (Default: table).

---

## Technical Notes

### PID Tracking
GOSP creates `.pid` files in `~/.gosp/` for every running service. The `stop` and `status` commands use these files to communicate with background processes without needing a heavy database.

### Log Locations
Logs are stored per profile in `~/.gosp/logs/`.
- `master_<name>.log`
- `worker_<id>.log`
