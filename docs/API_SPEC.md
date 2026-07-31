# API Specification - PrefixOS

PrefixOS exposes gRPC Unary/Streaming endpoints, REST HTTP endpoints, and Operational Admin endpoints.

---

## 1. gRPC API (`proto/v1/prefixos.proto`)

### 1.1 `MatchPrefix` Endpoint
Searches the Radix Tree for the longest matching sequence of tokens.

- **RPC**: `MatchPrefix(MatchPrefixRequest) returns (MatchPrefixResponse)`
- **Request**:
  - `tokens` (`repeated int32`): Token sequence IDs.
  - `tenant_id` (`string`): Multi-tenant isolation context.
- **Response**:
  - `matched_length` (`int32`): Number of matched prefix tokens.
  - `block_ids` (`repeated int32`): Physical block handles corresponding to matched prefix.

### 1.2 `Insert` Endpoint
Caches a new token sequence and allocates physical slab blocks for the non-matching suffix.

- **RPC**: `Insert(InsertRequest) returns (InsertResponse)`
- **Request**:
  - `tokens` (`repeated int32`): Sequence of token IDs.
  - `tenant_id` (`string`): Multi-tenant context.
- **Response**:
  - `success` (`bool`): Operation status.
  - `allocated_blocks` (`repeated int32`): Array of newly assigned physical block handles.

---

## 2. REST HTTP API

| Endpoint | Method | Description | Request Body | Response Body |
| :--- | :--- | :--- | :--- | :--- |
| `/v1/cache/match` | `POST` | Match token prefix | `{"tokens": [101, 205, 301]}` | `{"matched_length": 3, "block_ids": [12, 13]}` |
| `/v1/cache/store` | `POST` | Store token sequence | `{"tokens": [101, 205, 404]}` | `{"success": true, "block_ids": [14]}` |
| `/v1/cache/evict` | `POST` | Trigger explicit eviction | `{"count": 10}` | `{"freed_blocks": 10}` |
| `/metrics` | `GET` | Prometheus metrics | None | Prometheus text payload |
| `/health` | `GET` | Health check endpoint | None | `{"status": "UP"}` |

---

## 3. Operational Admin APIs (`/admin/*`)

| Admin Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/admin/snapshot` | `POST` | Trigger manual Copy-On-Write snapshot creation |
| `/admin/compact` | `POST` | Trigger manual slab memory defragmentation |
| `/admin/rebalance` | `POST` | Trigger cluster consistent hash ring rebalancing |
| `/admin/shards` | `GET` | Fetch shard status, node counts, and contention stats |
| `/admin/allocator` | `GET` | Fetch Slab Allocator metrics and fragmentation ratio |
| `/admin/reload` | `POST` | Reload configuration settings dynamically |
