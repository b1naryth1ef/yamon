package common

import "time"

type LogEntry struct {
	Time    time.Time         `json:"t"`
	Host    string            `json:"h"`
	Service string            `json:"s"`
	Level   string            `json:"l"`
	Data    string            `json:"d"`
	Tags    map[string]string `json:"g"`
}

func NewLogEntry(service, data string, tags map[string]string) *LogEntry {
	if tags == nil {
		tags = map[string]string{}
	}
	return &LogEntry{
		Time:    time.Now(),
		Service: service,
		Level:   "",
		Data:    data,
		Tags:    tags,
	}
}
