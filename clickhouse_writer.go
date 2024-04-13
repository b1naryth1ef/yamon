package yamon

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/b1naryth1ef/yamon/common"
)

type ClickhouseWriter struct {
	sync.Mutex

	cfg     common.ServerClickhouseConfig
	flushCh chan struct{}
	conn    driver.Conn
	batch   *ForwardBatch
}

func NewClickhouseWriter(cfg common.ServerClickhouseConfig) *ClickhouseWriter {
	return &ClickhouseWriter{
		cfg:     cfg,
		flushCh: make(chan struct{}),
		batch:   NewForwardBatch(),
	}
}

func makeMetricBatch(conn driver.Conn) (driver.Batch, error) {
	batch, err := conn.PrepareBatch(context.Background(), "")
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (m *ClickhouseWriter) writeLogs(batch *ForwardBatch) error {
	logBatch, err := m.conn.PrepareBatch(context.Background(), "INSERT INTO logs (when, host, service, level, data, tags)")
	if err != nil {
		return err
	}

	for _, logEntry := range batch.Logs {
		err = logBatch.Append(logEntry.Time, logEntry.Host, logEntry.Service, logEntry.Level, logEntry.Data, logEntry.Tags)
		if err != nil {
			return err
		}
	}

	err = logBatch.Send()
	if err != nil {
		return err
	}

	return nil
}

func (m *ClickhouseWriter) writeMetrics(batch *ForwardBatch) error {
	metricBatch, err := m.conn.PrepareBatch(context.Background(), "INSERT INTO metrics (when, type, host, name, value, tags)")
	if err != nil {
		return err
	}

	for _, metric := range batch.Metrics {
		err = metricBatch.Append(metric.Time, metric.Type, metric.Host, metric.Name, metric.Value, metric.Tags)
		if err != nil {
			return err
		}
	}

	err = metricBatch.Send()
	if err != nil {
		return err
	}

	return nil
}

func (m *ClickhouseWriter) flush() error {
	m.Lock()
	batch := m.batch
	m.batch = NewForwardBatch()
	m.Unlock()

	if m.conn == nil {
		conn, err := m.open()
		if err != nil {
			return err
		}
		m.conn = conn
	}

	var logError, metricError error
	if len(batch.Logs) > 0 {
		logError = m.writeLogs(batch)
		if logError != nil {
			ingestedLogs.WithLabelValues("dropped").Add(float64(len(batch.Logs)))
		} else {
			ingestedLogs.WithLabelValues("written").Add(float64(len(batch.Logs)))
		}
	}

	if len(batch.Metrics) > 0 {
		metricError = m.writeMetrics(batch)
		if metricError != nil {
			ingestedMetrics.WithLabelValues("dropped").Add(float64(len(batch.Metrics)))
		} else {
			ingestedMetrics.WithLabelValues("written").Add(float64(len(batch.Metrics)))
		}
	}

	if logError != nil {
		return logError
	} else if metricError != nil {
		return metricError
	}

	return nil
}

func orStr(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

func (m *ClickhouseWriter) open() (driver.Conn, error) {
	return clickhouse.Open(&clickhouse.Options{
		Addr: m.cfg.Targets,
		Auth: clickhouse.Auth{
			Database: m.cfg.Database,
			Username: m.cfg.Username,
			Password: m.cfg.Password,
		},
		Settings: map[string]any{
			"async_insert": 1,
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "yamon-server", Version: "0.1"},
			},
		},
		Debugf: func(format string, v ...interface{}) {
			fmt.Printf(format, v)
		},
	})
}

func (m *ClickhouseWriter) plsFlush() {
	select {
	case m.flushCh <- struct{}{}:
	default:
	}
}

func (m *ClickhouseWriter) Run() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-m.flushCh:
		case <-ticker.C:
		}

		err := m.flush()
		if err != nil {
			log.Printf("[ClickhouseWriter] error flushing: %v", err)
		}
	}
}

func (m *ClickhouseWriter) WriteMetrics(metric []*common.Metric) error {
	m.Lock()
	m.batch.Metrics = append(m.batch.Metrics, metric...)
	if len(m.batch.Metrics) > 5000 {
		m.plsFlush()
	}
	m.Unlock()
	return nil
}

func (m *ClickhouseWriter) WriteLogEntries(entries []*common.LogEntry) error {
	m.Lock()
	m.batch.Logs = append(m.batch.Logs, entries...)
	if len(m.batch.Logs) > 5000 {
		m.plsFlush()
	}
	m.Unlock()
	return nil
}
