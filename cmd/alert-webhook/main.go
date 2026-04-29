package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

type alertmanagerPayload struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	Alerts            []alertItem       `json:"alerts"`
}

type alertItem struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

func main() {
	port := os.Getenv("ALERT_WEBHOOK_PORT")
	if port == "" {
		port = "18080"
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
		})
	})

	mux.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
				"error": "method not allowed",
			})
			return
		}

		var payload alertmanagerPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Error("alert_webhook_decode_failed", "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error": "invalid json payload",
			})
			return
		}

		logger.Info(
			"alert_batch_received",
			"receiver", payload.Receiver,
			"status", payload.Status,
			"alerts", len(payload.Alerts),
		)

		for _, alert := range payload.Alerts {
			logger.Info(
				"alert_received",
				"alertname", alert.Labels["alertname"],
				"status", alert.Status,
				"severity", alert.Labels["severity"],
				"category", alert.Labels["category"],
				"service", firstNonEmpty(alert.Labels["service"], alert.Labels["job"]),
				"summary", alert.Annotations["summary"],
				"description", alert.Annotations["description"],
				"starts_at", alert.StartsAt,
				"ends_at", alert.EndsAt,
			)
		}

		writeJSON(w, http.StatusAccepted, map[string]any{
			"status": "accepted",
			"alerts": len(payload.Alerts),
		})
	})

	logger.Info("alert_webhook_started", "port", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Error("alert_webhook_stopped", "error", err)
		os.Exit(1)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}