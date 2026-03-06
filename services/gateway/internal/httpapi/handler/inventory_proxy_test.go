package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"finpharm-ai/services/gateway/internal/httpapi/handler"
	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

func TestGatewayInventory_ListMedicines_ProxyOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/medicines" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"items":[],"limit":10,"offset":0,"total":0},"request_id":"inv"}`))
	}))
	defer upstream.Close()

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())

	proxy := handler.NewInventoryProxyHandler(upstream.URL)
	r.GET("/v1/medicines", proxy.ListMedicines)

	req := httptest.NewRequest(http.MethodGet, "/v1/medicines?limit=1&offset=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestGatewayInventory_GetMedicine_ProxyOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/medicines/PARA500" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"id":"PARA500","name":"Paracetamol 500mg","type":"OTC"},"request_id":"inv"}`))
	}))
	defer upstream.Close()

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())

	proxy := handler.NewInventoryProxyHandler(upstream.URL)
	r.GET("/v1/medicines/:id", proxy.GetMedicine)

	req := httptest.NewRequest(http.MethodGet, "/v1/medicines/PARA500", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}