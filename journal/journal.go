package journal

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/b1naryth1ef/yamon/common"
	"github.com/coreos/go-systemd/sdjournal"
)

func getLevelName(level string) string {
	switch level {
	case "0", "1", "2":
		return "critical"
	case "3":
		return "error"
	case "4":
		return "warning"
	case "5", "6":
		return "info"
	case "7":
		return "debug"
	default:
		return ""
	}
}

type JournalTracker interface {
	CommitCursor(cursor string) error
	LastCursor() (string, error)
}

type FileBasedJournalTracker struct {
	file  *os.File
	sync  int
	count int
}

func NewFileBasedJournalTracker(path string, sync int) (*FileBasedJournalTracker, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &FileBasedJournalTracker{file: f, sync: sync}, nil
}

func (f *FileBasedJournalTracker) CommitCursor(cursor string) error {
	_, err := f.file.WriteAt([]byte(cursor), 0)
	if err != nil {
		return err
	}
	f.count += 1
	if f.sync > 0 && f.count%f.sync == 0 {
		return f.file.Sync()
	}
	return nil
}

func (f *FileBasedJournalTracker) LastCursor() (string, error) {
	_, err := f.file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	data := make([]byte, 256)
	n, err := f.file.Read(data)

	if err == io.EOF && n == 0 {
		return "", nil
	}

	if err != nil {
		return "", err
	}
	return string(data[:n]), nil
}

func (f *FileBasedJournalTracker) Close() {
	f.file.Close()
}

type JournalReader struct {
	sink    common.LogSink
	tracker JournalTracker
}

func NewJournalReader(sink common.LogSink, tracker JournalTracker) *JournalReader {
	return &JournalReader{sink: sink, tracker: tracker}
}

func (j *JournalReader) Run() error {
	data, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		log.Printf("no machine id")
		return err
	}
	machineId := strings.TrimSpace(string(data))

	journ, err := sdjournal.NewJournalFromDir(filepath.Join("/var/log/journal", machineId))
	if err != nil {
		log.Printf("Failed to open journal: %v", err)
		return err
	}

	if j.tracker != nil {
		cursor, err := j.tracker.LastCursor()
		if err != nil {
			return err
		}

		log.Printf("Recovering w/ cursor %v", cursor)

		if cursor != "" {
			err = journ.SeekCursor(cursor)
		} else {
			err = journ.SeekTail()
		}
		journ.Next()
	} else {
		err = journ.SeekTail()
		if err != nil {
			log.Printf("Failed to seek journal: %v", err)
			return err
		}
	}

	journ.Previous()
	journ.Next()

	for {
		for {
			wait := journ.Wait(sdjournal.IndefiniteWait)
			if wait < 0 {
				return nil
			} else if wait == sdjournal.SD_JOURNAL_NOP {
				continue
			} else {
				break
			}
		}

		var nextId uint64 = 1
		for nextId > 0 {
			var err error
			nextId, err = journ.Next()
			if err != nil {
				log.Printf("Failed to next: %v", err)
				return err
			}

			if nextId == 0 {
				continue
			}

			entry, err := journ.GetEntry()
			if err != nil {
				log.Printf("Failed to get entry: %v", err)
				return err
			}

			message := entry.Fields["MESSAGE"]
			delete(entry.Fields, "MESSAGE")
			delete(entry.Fields, "_HOSTNAME")
			delete(entry.Fields, "_SYSTEMD_INVOCATION_ID")
			delete(entry.Fields, "_STREAM_ID")

			logEntry := common.NewLogEntry(entry.Fields["SYSLOG_IDENTIFIER"], message, entry.Fields)
			logEntry.Time = time.Unix(int64(entry.RealtimeTimestamp/1000000), 0)
			logEntry.Level = getLevelName(entry.Fields["PRIORITY"])
			j.sink.WriteLog(logEntry)

			cursor, err := journ.GetCursor()
			if err != nil {
				return err
			}
			err = j.tracker.CommitCursor(cursor)
			if err != nil {
				return err
			}
		}
	}
}
