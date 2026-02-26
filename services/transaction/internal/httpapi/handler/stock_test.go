package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"finpharm-ai/services/transaction/internal/config"
	"finpharm-ai/services/transaction/internal/httpapi"
	"github.com/gin-gonic/gin"
)

func TestCheckStock_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock Inventory Service
	inv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/stock/check" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"medicine_id":"PARA500","requested_qty":10,"available_qty":80,"is_available":true},"request_id":"inv-1"}`))
	}))
	defer inv.Close()

	r := httpapi.NewRouter(config.Config{
		AppEnv:            "local",
		InventoryBaseURL:  inv.URL,
	})

	body := []byte(`{"medicine_id":"PARA500","qty":10}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// memastikan envelope success masih ada
	if !contains(w.Body.String(), `"data"`) || !contains(w.Body.String(), `"request_id"`) {
		t.Fatalf("expected envelope data+request_id, got: %s", w.Body.String())
	}
}

func TestCheckStock_MedicineNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock Inventory Service -> return 404 for unknown medicine
	inv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/stock/check" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":"MEDICINE_NOT_FOUND","message":"medicine not found","details":{"resource":"medicine","key":"PARA200"},"request_id":"inv-404"}}`))
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

func TestCheckStock_UsecaseValidation_MedicineIDEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Inventory tidak penting karena validasi gagal sebelum call inventory
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

	if !contains(w.Body.String(), `"field":"medicine_id"`) {
		t.Fatalf("expected field medicine_id in details, got: %s", w.Body.String())
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

	if !contains(w.Body.String(), `"request_id"`) {
		t.Fatalf("expected response body to contain request_id, got: %s", w.Body.String())
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}