package common

type Batch struct {
	Metrics []*Metric   `json:"m"`
	Logs    []*LogEntry `json:"l"`
}

func NewBatch() *Batch {
	return &Batch{
		Metrics: make([]*Metric, 0),
		Logs:    make([]*LogEntry, 0),
	}
}
