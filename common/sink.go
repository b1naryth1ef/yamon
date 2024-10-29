package common

type MetricSink interface {
	WriteMetric(*Metric)
}

type LogSink interface {
	WriteLog(*LogEntry)
}

type EventSink interface {
	WriteEvent(*Event)
}

type Sink interface {
	MetricSink
	LogSink
	EventSink
}
