package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"prefixos/internal/memory"
	"prefixos/internal/radix"
)

func TestHTTPServer_HealthAndMetrics(t *testing.T) {
	bm := memory.NewBlockManager(100)
	tree := radix.NewTree(bm)
	httpServer := NewHTTPServer("127.0.0.1:0", tree, bm)

	// Test /health
	reqHealth := httptest.NewRequest("GET", "/health", nil)
	wHealth := httptest.NewRecorder()
	httpServer.server.Handler.ServeHTTP(wHealth, reqHealth)

	if wHealth.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /health, got %d", wHealth.Code)
	}

	// Test /metrics
	reqMetrics := httptest.NewRequest("GET", "/metrics", nil)
	wMetrics := httptest.NewRecorder()
	httpServer.server.Handler.ServeHTTP(wMetrics, reqMetrics)

	if wMetrics.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /metrics, got %d", wMetrics.Code)
	}
}

func TestHTTPServer_CacheStoreAndMatch(t *testing.T) {
	bm := memory.NewBlockManager(100)
	tree := radix.NewTree(bm)
	httpServer := NewHTTPServer("127.0.0.1:0", tree, bm)

	// Store sequence via POST /v1/cache/store
	storeBody, _ := json.Marshal(map[string]interface{}{
		"tokens": []int32{100, 200, 300},
	})
	reqStore := httptest.NewRequest("POST", "/v1/cache/store", bytes.NewBuffer(storeBody))
	wStore := httptest.NewRecorder()
	httpServer.server.Handler.ServeHTTP(wStore, reqStore)

	if wStore.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /v1/cache/store, got %d", wStore.Code)
	}

	var storeResp storeResponse
	_ = json.NewDecoder(wStore.Body).Decode(&storeResp)
	if !storeResp.Success || len(storeResp.BlockIDs) == 0 {
		t.Fatalf("expected store success, got %+v", storeResp)
	}

	// Match sequence via POST /v1/cache/match
	matchBody, _ := json.Marshal(map[string]interface{}{
		"tokens": []int32{100, 200, 300, 400},
	})
	reqMatch := httptest.NewRequest("POST", "/v1/cache/match", bytes.NewBuffer(matchBody))
	wMatch := httptest.NewRecorder()
	httpServer.server.Handler.ServeHTTP(wMatch, reqMatch)

	if wMatch.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /v1/cache/match, got %d", wMatch.Code)
	}

	var matchResp matchResponse
	_ = json.NewDecoder(wMatch.Body).Decode(&matchResp)
	if matchResp.MatchedLength != 3 {
		t.Fatalf("expected matched length 3, got %d", matchResp.MatchedLength)
	}
}
