package yamon

import (
	"github.com/b1naryth1ef/yamon/common"
)

type DataWriter interface {
	WriteMetrics([]*common.Metric) error
	WriteLogEntries([]*common.LogEntry) error
	WriteEvents([]*common.Event) error
}

type SinkMetadataFilter struct {
	hostname string
	tags     map[string]string
	sink     common.Sink
}

func NewSinkMetadataFilter(hostname string, tags map[string]string, sink common.Sink) *SinkMetadataFilter {
	return &SinkMetadataFilter{
		hostname: hostname,
		tags:     tags,
		sink:     sink,
	}
}

func (s *SinkMetadataFilter) WriteMetric(metric *common.Metric) {
	if s.tags != nil {
		for k, v := range s.tags {
			metric.Tags[k] = v
		}
	}
	metric.Host = s.hostname
	s.sink.WriteMetric(metric)
}

func (s *SinkMetadataFilter) WriteLog(log *common.LogEntry) {
	if s.tags != nil {
		for k, v := range s.tags {
			log.Tags[k] = v
		}
	}
	log.Host = s.hostname
	s.sink.WriteLog(log)
}

func (s *SinkMetadataFilter) WriteEvent(event *common.Event) {
	if s.tags != nil {
		for k, v := range s.tags {
			event.Tags[k] = v
		}
	}
	event.Host = s.hostname
	s.sink.WriteEvent(event)
}
