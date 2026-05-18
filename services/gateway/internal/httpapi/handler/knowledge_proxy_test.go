package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

func TestKnowledgeProxyHandler_ChatSOP_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/sop" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"question":"apa edukasi minimal untuk paracetamol otc?"`) {
			t.Fatalf("unexpected upstream body: %s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"question":"apa edukasi minimal untuk paracetamol otc?","answer":"Edukasi minimal wajib disampaikan [S1].","fallback":false,"citations":["[S1]"],"sources":[{"ref":"[S1]","title":"SOP Penjualan Obat OTC Paracetamol 500mg","category":"otc-sales","source_key":"otc-paracetamol.md","heading":"Edukasi Minimal","score":0.79}],"confidence":{"top_score":0.79,"min_top_score":0.62,"retrieved_count":2,"used_source_count":1}},"request_id":"req-123"}`))
	}))
	defer upstream.Close()

	router := gin.New()
	router.Use(middleware.RequestID())
	h := NewKnowledgeProxyHandler(upstream.URL)
	router.POST("/v1/chat/sop", h.ChatSOP)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/sop", strings.NewReader(`{"question":"apa edukasi minimal untuk paracetamol otc?","top_k":5,"min_score":0.45}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"fallback":false`) {
		t.Fatalf("expected fallback=false, got body=%s", rec.Body.String())
	}
}

func TestKnowledgeProxyHandler_ChatSOP_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.RequestID())
	h := NewKnowledgeProxyHandler("http://localhost:8084")
	router.POST("/v1/chat/sop", h.ChatSOP)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/sop", strings.NewReader(`{"question":"","top_k":5,"min_score":0.45}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"VALIDATION_ERROR"`) {
		t.Fatalf("expected validation error, got body=%s", rec.Body.String())
	}
}

func TestKnowledgeProxyHandler_ChatSOP_UpstreamError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.RequestID())
	h := NewKnowledgeProxyHandler("http://127.0.0.1:1")
	router.POST("/v1/chat/sop", h.ChatSOP)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/sop", strings.NewReader(`{"question":"apakah amoxicillin bisa dijual tanpa resep?","top_k":5,"min_score":0.45}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"UPSTREAM_ERROR"`) {
		t.Fatalf("expected UPSTREAM_ERROR, got body=%s", rec.Body.String())
	}
}