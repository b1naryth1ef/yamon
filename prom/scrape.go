package prom

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/b1naryth1ef/yamon/common"
	"github.com/prometheus/common/expfmt"
)

type Scraper struct {
	config   common.PrometheusScraperConfig
	sink     common.MetricSink
	interval time.Duration

	http http.Client
}

func NewScraper(config common.PrometheusScraperConfig, sink common.MetricSink) (*Scraper, error) {
	interval, err := time.ParseDuration(config.Interval)
	if err != nil {
		return nil, err
	}

	var timeout time.Duration
	if config.Timeout != "" {
		timeout, err = time.ParseDuration(config.Timeout)
		if err != nil {
			return nil, err
		}
	} else {
		timeout = time.Second * 5
	}

	return &Scraper{
		config:   config,
		interval: interval,
		sink:     sink,
		http: http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (s *Scraper) Run() {
	for {
		s.scrape()
		time.Sleep(s.interval)
	}
}

func (s *Scraper) scrape() {
	res, err := s.http.Get(s.config.URL)
	if err != nil {
		slog.Error("failed to prom scrape", slog.String("url", s.config.URL), slog.Any("error", err))
		return
	}
	defer res.Body.Close()

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(res.Body)
	if err != nil {
		slog.Error("failed to parse prom metric data", slog.String("url", s.config.URL), slog.Any("error", err))
		return
	}

	for _, metricFamily := range mf {
		for _, metric := range metricFamily.Metric {
			tags := map[string]string{}
			for _, label := range metric.Label {
				tags[label.GetName()] = label.GetValue()
			}

			name := metricFamily.GetName()
			if s.config.Prefix != "" {
				name = s.config.Prefix + name
			}

			if metric.Gauge != nil {
				value := metric.Gauge.GetValue()
				s.sink.WriteMetric(common.NewGauge(name, value, tags))
			} else if metric.Counter != nil {
				value := metric.Counter.GetValue()
				s.sink.WriteMetric(common.NewCounter(name, value, tags))
			} else {
				slog.Debug("skipping unsupported prom metric type", slog.String("name", *metricFamily.Name), slog.Any("type", metricFamily.Type))
			}
		}
	}
}
