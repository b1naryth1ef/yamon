package common

import (
	"time"

	"golang.org/x/exp/constraints"
)

type MetricType = string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter            = "counter"
)

type Metric struct {
	Time  time.Time         `json:"t"`
	Type  MetricType        `json:"m"`
	Host  string            `json:"h"`
	Name  string            `json:"n"`
	Value float64           `json:"v"`
	Tags  map[string]string `json:"g"`
}

func NewGauge[T constraints.Integer | constraints.Float](name string, value T, tags map[string]string) *Metric {
	return NewMetric(name, MetricTypeGauge, value, tags)
}

func NewCounter[T constraints.Integer | constraints.Float](name string, value T, tags map[string]string) *Metric {
	return NewMetric(name, MetricTypeCounter, value, tags)
}

func NewMetric[T constraints.Integer | constraints.Float](name string, mtype MetricType, value T, tags map[string]string) *Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	return &Metric{
		Time:  time.Now(),
		Type:  mtype,
		Name:  name,
		Value: float64(value),
		Tags:  tags,
	}
}
