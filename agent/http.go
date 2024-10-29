package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/alioygur/gores"
	"github.com/b1naryth1ef/yamon/common"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type AgentHTTPServer struct {
	sink common.Sink
}

func NewAgentHTTPServer(sink common.Sink) *AgentHTTPServer {
	return &AgentHTTPServer{sink: sink}
}

func (a *AgentHTTPServer) Run(bind string) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Handle("/metrics", promhttp.Handler())

	r.Post("/v1/data", a.postData)
	r.Post("/v1/webhook", a.postWebhook)

	return http.ListenAndServe(bind, r)
}

type PostDataRequest struct {
	Metrics []common.Metric   `json:"metrics,omitempty"`
	Events  []common.Event    `json:"events,omitempty"`
	Logs    []common.LogEntry `json:"logs,omitempty"`
}

func (a *AgentHTTPServer) postData(w http.ResponseWriter, r *http.Request) {
	var data PostDataRequest

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		gores.Error(w, http.StatusBadRequest, fmt.Sprintf("invalid json: %v", err))
		return
	}

	if data.Metrics != nil {
		for idx := range data.Metrics {
			a.sink.WriteMetric(&data.Metrics[idx])
		}
	}

	if data.Events != nil {
		for idx := range data.Events {
			a.sink.WriteEvent(&data.Events[idx])
		}
	}

	if data.Logs != nil {
		for idx := range data.Logs {
			a.sink.WriteLog(&data.Logs[idx])
		}
	}

	gores.NoContent(w)
}

func (a *AgentHTTPServer) postWebhook(w http.ResponseWriter, r *http.Request) {
	tags := map[string]string{}
	data := map[string]any{}

	tags["remote-addr"] = r.RemoteAddr
	tags["content-type"] = r.Header.Get("Content-Type")

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		r.ParseMultipartForm(2048)

		for k, v := range r.Form {
			var vJSON any
			err := json.Unmarshal([]byte(v[0]), &vJSON)
			if err == nil {
				data[k] = vJSON
			} else {
				data[k] = v[0]
			}
		}
	}

	a.sink.WriteEvent(common.NewEventJSON("yamon-agent.webhook", data, tags))
	gores.NoContent(w)
}
