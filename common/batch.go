package common

type Batch struct {
	Metrics []*Metric   `json:"m"`
	Logs    []*LogEntry `json:"l"`
	Events  []*Event    `json:"e"`
}

func NewBatch() *Batch {
	return &Batch{
		Metrics: make([]*Metric, 0),
		Logs:    make([]*LogEntry, 0),
		Events:  make([]*Event, 0),
	}
}
