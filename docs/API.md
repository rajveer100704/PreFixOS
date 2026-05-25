# gRPC API Reference

RadixKV exposes its Control Plane via a gRPC interface (`proto/kvcache.proto`). 

Inference engines (acting as the Data Plane) call these endpoints to determine where to find or place physical memory for LLM attention weights.

---

## `MatchPrefix`
Searches the Radix Tree for the longest shared sequence of tokens.

**Request (`MatchPrefixRequest`)**
- `tokens` (`repeated int32`): The sequence of token IDs to match against the cache.

**Response (`MatchPrefixResponse`)**
- `matched_length` (`int32`): How many sequential tokens successfully matched a shared prefix.
- `block_ids` (`repeated int32`): An ordered array of physical Block IDs. The inference engine uses these to index into shared memory/VRAM to load the cached attention tensors.

---

## `Insert`
Registers a new sequence of tokens into the Radix Tree, allocating new memory blocks for the un-cached suffix, and splitting nodes if necessary.

**Request (`InsertRequest`)**
- `tokens` (`repeated int32`): The full sequence of token IDs to cache (e.g., prompt + generated tokens).

**Response (`InsertResponse`)**
- `success` (`bool`): `true` if the tokens were successfully cached. `false` if the engine ran out of memory (OOM).
- `block_ids` (`repeated int32`): An ordered array of newly allocated physical Block IDs. The inference engine should write its newly computed attention tensors to the corresponding VRAM offsets for these blocks.
