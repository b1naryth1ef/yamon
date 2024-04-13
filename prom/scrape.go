package prom

import (
	"log"
	"net/http"
	"time"

	"github.com/b1naryth1ef/yamon/common"
	"github.com/prometheus/common/expfmt"
)

type Scraper struct {
	url      string
	sink     common.MetricSink
	interval time.Duration

	http http.Client
}

func NewScraper(url string, interval time.Duration, sink common.MetricSink) *Scraper {
	return &Scraper{
		url:      url,
		sink:     sink,
		interval: interval,
	}
}

func (s *Scraper) Run() {
	for {
		s.scrape()
		time.Sleep(s.interval)
	}
}

func (s *Scraper) scrape() {
	res, err := s.http.Get(s.url)
	if err != nil {
		log.Printf("Failed to scrape (%v): %v", s.url, err)
		return
	}
	defer res.Body.Close()

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(res.Body)
	if err != nil {
		log.Printf("Failed to parse metric families (%v): %v", s.url, err)
		return
	}

	for _, metricFamily := range mf {
		for _, metric := range metricFamily.Metric {
			tags := map[string]string{}
			for _, label := range metric.Label {
				tags[label.GetName()] = label.GetValue()
			}

			if metric.Gauge != nil {
				value := metric.Gauge.GetValue()
				s.sink.WriteMetric(common.NewGauge(*metricFamily.Name, value, tags))
			} else if metric.Counter != nil {
				value := metric.Counter.GetValue()
				s.sink.WriteMetric(common.NewCounter(*metricFamily.Name, value, tags))
			} else {
				// log.Printf("WARNING: unsupported metric type for '%v' (%v)", *metricFamily.Name, metricFamily.Type)
			}
		}
	}
}
