package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"prefixos/internal/memory"
	"prefixos/internal/radix"
)

// HTTPServer handles REST endpoints, operational metrics, and health checks.
type HTTPServer struct {
	tree        *radix.Tree
	bm          *memory.BlockManager
	addr        string
	server      *http.Server
	lookupCount uint64
	storeCount  uint64
}

// NewHTTPServer initializes the REST and administrative endpoint server.
func NewHTTPServer(addr string, tree *radix.Tree, bm *memory.BlockManager) *HTTPServer {
	h := &HTTPServer{
		tree: tree,
		bm:   bm,
		addr: addr,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/metrics", h.handleMetrics)
	mux.HandleFunc("/v1/stats", h.handleStats)
	mux.HandleFunc("/v1/cache/match", h.handleCacheMatch)
	mux.HandleFunc("/v1/cache/store", h.handleCacheStore)

	h.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return h
}

// Start begins listening on the configured HTTP address.
func (h *HTTPServer) Start() error {
	return h.server.ListenAndServe()
}

// Stop gracefully shuts down the HTTP server.
func (h *HTTPServer) Stop() error {
	return h.server.Close()
}

func (h *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UP"})
}

func (h *HTTPServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	stats := h.bm.Stats()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP prefixos_memory_total_blocks Total capacity of memory blocks\n")
	fmt.Fprintf(w, "# TYPE prefixos_memory_total_blocks gauge\n")
	fmt.Fprintf(w, "prefixos_memory_total_blocks %d\n", stats.TotalBlocks)
	fmt.Fprintf(w, "# HELP prefixos_memory_allocated_blocks Allocated memory blocks count\n")
	fmt.Fprintf(w, "# TYPE prefixos_memory_allocated_blocks gauge\n")
	fmt.Fprintf(w, "prefixos_memory_allocated_blocks %d\n", stats.AllocatedBlocks)
	fmt.Fprintf(w, "# HELP prefixos_memory_free_blocks Free memory blocks count\n")
	fmt.Fprintf(w, "# TYPE prefixos_memory_free_blocks gauge\n")
	fmt.Fprintf(w, "prefixos_memory_free_blocks %d\n", stats.FreeBlocks)
	fmt.Fprintf(w, "# HELP prefixos_memory_fragmentation_ratio Memory fragmentation ratio\n")
	fmt.Fprintf(w, "# TYPE prefixos_memory_fragmentation_ratio gauge\n")
	fmt.Fprintf(w, "prefixos_memory_fragmentation_ratio %.4f\n", stats.Fragmentation)
}

func (h *HTTPServer) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := h.bm.Stats()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"total_blocks":     stats.TotalBlocks,
		"allocated_blocks": stats.AllocatedBlocks,
		"free_blocks":      stats.FreeBlocks,
		"fragmentation":    stats.Fragmentation,
		"lookup_count":     atomic.LoadUint64(&h.lookupCount),
		"store_count":      atomic.LoadUint64(&h.storeCount),
	})
}

type matchRequest struct {
	Tokens []int32 `json:"tokens"`
}

type matchResponse struct {
	MatchedLength int     `json:"matched_length"`
	BlockIDs      []int32 `json:"block_ids"`
}

func (h *HTTPServer) handleCacheMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req matchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	atomic.AddUint64(&h.lookupCount, 1)
	matchedLen, blockIDs := h.tree.FindLongestPrefix(req.Tokens)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(matchResponse{
		MatchedLength: matchedLen,
		BlockIDs:      blockIDs,
	})
}

type storeRequest struct {
	Tokens []int32 `json:"tokens"`
}

type storeResponse struct {
	Success  bool    `json:"success"`
	BlockIDs []int32 `json:"block_ids"`
}

func (h *HTTPServer) handleCacheStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req storeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	atomic.AddUint64(&h.storeCount, 1)
	success, blockIDs := h.tree.InsertTokens(req.Tokens)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(storeResponse{
		Success:  success,
		BlockIDs: blockIDs,
	})
}
