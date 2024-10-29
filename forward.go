package yamon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/b1naryth1ef/yamon/common"
)

type ForwardClient struct {
	target string
	auth   string
	client http.Client
}

// NewForwardClient creates a client pointing to a target like tcp://localhost:6691
func NewForwardClient(target string) (*ForwardClient, error) {
	auth := "none"
	url, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	if url.User != nil {
		auth = url.User.Username()
		pw, ok := url.User.Password()
		if !ok {
			return nil, fmt.Errorf("url expected both user and password")
		}
		auth = auth + ":" + pw
	}

	url.User = nil

	return &ForwardClient{target: url.String(), auth: auth}, nil
}

func (f *ForwardClient) SubmitBatch(batch *common.Batch) error {
	data, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", f.target+"/v1/submit-batch", bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", f.auth)

	res, err := f.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 204 {
		return fmt.Errorf("invalid status code %d", res.StatusCode)
	}

	return nil
}

type ForwardClientSinkFlushConfig struct {
	MetricThreshold uint
	LogThreshold    uint
	EventThreshold  uint
	Interval        time.Duration
}

type ForwardClientSink struct {
	sync.Mutex

	flushCfg ForwardClientSinkFlushConfig
	flushCh  chan struct{}
	client   *ForwardClient
	batch    *common.Batch
}

func NewForwardClientSink(client *ForwardClient, flushCfg ForwardClientSinkFlushConfig) *ForwardClientSink {
	sink := &ForwardClientSink{
		flushCfg: flushCfg,
		flushCh:  make(chan struct{}),
		client:   client,
		batch:    common.NewBatch(),
	}
	go sink.run()
	return sink
}

func (f *ForwardClientSink) flush() error {
	f.Lock()
	batch := f.batch
	f.batch = common.NewBatch()
	f.Unlock()

	// TODO: retries here, will block further flushes
	return f.client.SubmitBatch(batch)
}

func (f *ForwardClientSink) plsFlush() {
	select {
	case f.flushCh <- struct{}{}:
	default:
	}
}

func (f *ForwardClientSink) run() {
	timer := time.NewTicker(time.Second * f.flushCfg.Interval)
	for {
		select {
		case <-f.flushCh:
		case <-timer.C:
		}

		err := f.flush()
		if err != nil {
			slog.Error("failed to flush data from ForwardClientSink", slog.Any("error", err))
		}
	}
}

func (f *ForwardClientSink) WriteMetric(metric *common.Metric) {
	f.Lock()
	f.batch.Metrics = append(f.batch.Metrics, metric)

	if len(f.batch.Metrics) > int(f.flushCfg.MetricThreshold) {
		f.plsFlush()
	}

	f.Unlock()
}

func (f *ForwardClientSink) WriteLog(entry *common.LogEntry) {
	f.Lock()
	f.batch.Logs = append(f.batch.Logs, entry)
	if len(f.batch.Logs) > int(f.flushCfg.LogThreshold) {
		f.plsFlush()
	}
	f.Unlock()
}

func (f *ForwardClientSink) WriteEvent(event *common.Event) {
	f.Lock()
	f.batch.Events = append(f.batch.Events, event)
	if len(f.batch.Events) > int(f.flushCfg.EventThreshold) {
		f.plsFlush()
	}
	f.Unlock()
}
