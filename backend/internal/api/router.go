package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/MTPeraya/stampede-shield/backend/internal/features"
	"github.com/MTPeraya/stampede-shield/backend/internal/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(service *features.Service, m *metrics.Metrics) http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	r.Get("/features/{entity_id}", featureHandler(service))
	r.Post("/mode", modeHandler(service))
	r.Get("/metrics/stream", metricsStreamHandler(service, m))
	r.Post("/loadgen/burst", burstHandler(service))
	r.Handle("/metrics", promhttp.Handler())
	return r
}

func metricsStreamHandler(service *features.Service, m *metrics.Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		encoder := json.NewEncoder(w)
		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				_, _ = w.Write([]byte("data: "))
				if err := encoder.Encode(m.Snapshot(string(service.Mode()))); err != nil {
					return
				}
				_, _ = w.Write([]byte("\n"))
				flusher.Flush()
			}
		}
	}
}

func burstHandler(service *features.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rps := queryInt(r, "rps", 100)
		seconds := queryInt(r, "seconds", 10)
		keys := queryInt(r, "keys", 10)
		if rps < 1 || rps > 10000 || seconds < 1 || seconds > 300 || keys < 1 || keys > 1000000 {
			http.Error(w, "rps, seconds, and keys are out of range", http.StatusBadRequest)
			return
		}
		go service.RunBurst(rps, seconds, keys)
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("burst started\n"))
	}
}

func queryInt(r *http.Request, name string, fallback int) int {
	value, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil {
		return fallback
	}
	return value
}

func modeHandler(service *features.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mode := features.Mode(r.URL.Query().Get("value"))
		if mode != features.ModeNaive && mode != features.ModeHardened {
			http.Error(w, "value must be naive or hardened", http.StatusBadRequest)
			return
		}
		if err := service.SetMode(mode); err != nil {
			http.Error(w, "value must be naive or hardened", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func featureHandler(service *features.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer service.Observe(start)
		entityID, err := strconv.ParseInt(chi.URLParam(r, "entity_id"), 10, 64)
		if err != nil || entityID < 1 {
			http.Error(w, "entity_id must be a positive integer", http.StatusBadRequest)
			return
		}
		feature, err := service.Get(r.Context(), entityID)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				http.Error(w, "upstream timeout", http.StatusServiceUnavailable)
				return
			}
			http.Error(w, "feature unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(feature)
	}
}
