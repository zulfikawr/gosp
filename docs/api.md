# GOSP API & Metadata 🦞

The GOSP Master node provides a high-performance REST API designed to be a drop-in replacement for the **Brave Search API v1**.

## Primary Endpoint: `GET /web/search`

### Parameters
| Parameter | Default | Description |
| :--- | :--- | :--- |
| `q` | (None) | The search query string. |
| `engine` | `duckduckgo` | Target engine: `google`, `duckduckgo`, `brave`. |
| `count` | `10` | Number of results to return. |
| `metadata` | `false` | Enable opt-in OSP signals. |

## Metadata Mode (`?metadata=true`)

When requested, GOSP injects extra transparency signals into the response. This is unique to GOSP and allows for deep analysis of the distributed scrape.

### 1. `osp_signals` (Per Result)
Each result item includes:
- `source`: The specific engine that found the link.
- `worker_id`: The ID of the node that performed the scrape.
- `region`: The geographic location of the worker.

### 2. `osp_performance` (Root Level)
Provides a latency breakdown:
- `worker_scrape_ms`: Time taken by the node to fetch data.
- `master_agg_ms`: Time taken by the orchestrator to clean and deduplicate.

### 3. `osp_diagnostics` (On Error)
If a search fails in metadata mode, the response includes internal diagnostic data (Target Engine, Raw Error, Trace ID) to assist in cluster troubleshooting.
