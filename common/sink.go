package common

type MetricSink interface {
	WriteMetric(*Metric)
}

type LogSink interface {
	WriteLog(*LogEntry)
}

type Sink interface {
	MetricSink
	LogSink
}
