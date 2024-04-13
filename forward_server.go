package yamon

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alioygur/gores"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ForwardServer struct {
	w    DataWriter
	keys map[string]string
}

func NewForwardServer(w DataWriter, keys map[string]string) *ForwardServer {
	if len(keys) == 0 {
		keys = nil
	}
	return &ForwardServer{w: w, keys: keys}
}

func (f *ForwardServer) Run(bind string) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Handle("/metrics", promhttp.Handler())
	r.Post("/v1/submit-batch", f.submitBatch)

	return http.ListenAndServe(bind, r)
}

func (f *ForwardServer) submitBatch(w http.ResponseWriter, r *http.Request) {
	if f.keys != nil {
		auth := r.Header.Get("Authorization")
		parts := strings.Split(auth, ":")
		if len(parts) != 2 {
			gores.Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		key, ok := f.keys[parts[0]]
		if !ok {
			gores.Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if key != parts[1] {
			gores.Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}
	}

	var request ForwardBatch
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		gores.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	err = f.w.WriteMetrics(request.Metrics)
	if err != nil {
		gores.Error(w, http.StatusInternalServerError, "failed to write metrics")
		return
	}
	err = f.w.WriteLogEntries(request.Logs)
	if err != nil {
		gores.Error(w, http.StatusInternalServerError, "failed to write log entries")
		return
	}
	gores.NoContent(w)
}
