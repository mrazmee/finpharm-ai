package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"finpharm-ai/services/transaction/internal/config"
	"finpharm-ai/services/transaction/internal/httpapi"
	"github.com/gin-gonic/gin"
)

func TestCheckStock_OK_PropagatesRequestIDToInventory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	inv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/stock/check" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("X-Request-ID") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.Header.Get("X-Caller-Service") != "transaction" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"medicine_id":"PARA500","requested_qty":10,"available_qty":80,"is_available":true},"request_id":"inv-1"}`))
	}))
	defer inv.Close()

	r := httpapi.NewRouter(config.Config{
		AppEnv:           "local",
		InventoryBaseURL: inv.URL,
	})

	body := []byte(`{"medicine_id":"PARA500","qty":10}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "t-req-123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestCheckStock_MedicineNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	inv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer inv.Close()

	r := httpapi.NewRouter(config.Config{
		AppEnv:           "local",
		InventoryBaseURL: inv.URL,
	})

	body := []byte(`{"medicine_id":"PARA200","qty":10}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestCheckStock_InventoryTimeout_Returns502(t *testing.T) {
	gin.SetMode(gin.TestMode)

	inv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"medicine_id":"PARA500","requested_qty":1,"available_qty":80,"is_available":true},"request_id":"inv-slow"}`))
	}))
	defer inv.Close()

	r := httpapi.NewRouter(config.Config{
		AppEnv:           "local",
		InventoryBaseURL: inv.URL,
	})

	body := []byte(`{"medicine_id":"PARA500","qty":1}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	start := time.Now()
	r.ServeHTTP(w, req)
	elapsed := time.Since(start)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d, body=%s", w.Code, w.Body.String())
	}

	if elapsed > 2800*time.Millisecond {
		t.Fatalf("expected request to timeout earlier, elapsed=%v", elapsed)
	}
}

func TestCheckStock_UsecaseValidation_MedicineIDEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := httpapi.NewRouter(config.Config{
		AppEnv:           "local",
		InventoryBaseURL: "http://example.com",
	})

	body := []byte(`{"medicine_id":"   ","qty":1}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestCheckStock_InvalidBody_HasRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := httpapi.NewRouter(config.Config{
		AppEnv:           "local",
		InventoryBaseURL: "http://example.com",
	})

	body := []byte(`{"medicine_id":"","qty":0}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}

	if !bytes.Contains(w.Body.Bytes(), []byte(`"request_id"`)) {
		t.Fatalf("expected response body to contain request_id, got: %s", w.Body.String())
	}
}