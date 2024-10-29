package common

import (
	"encoding/json"
	"time"
)

type Event struct {
	Time time.Time         `json:"t"`
	Host string            `json:"h"`
	Type string            `json:"e"`
	Data string            `json:"d"`
	Tags map[string]string `json:"g"`
}

func NewEvent(eventType, data string, tags map[string]string) *Event {
	if tags == nil {
		tags = map[string]string{}
	}
	return &Event{
		Time: time.Now(),
		Type: eventType,
		Data: data,
		Tags: tags,
	}
}

func NewEventJSON[T any](eventType string, data T, tags map[string]string) *Event {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		dataBytes = []byte("<invalid json>")
	}

	if tags == nil {
		tags = map[string]string{}
	}
	return &Event{
		Time: time.Now(),
		Type: eventType,
		Data: string(dataBytes),
		Tags: tags,
	}
}
