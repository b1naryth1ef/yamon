package journal

import (
	"io"
	"os"
)

type JournalTracker interface {
	CommitCursor(cursor string) error
	LastCursor() (string, error)
}

type NoopJournalTracker struct {
}

func (n *NoopJournalTracker) CommitCursor(cursor string) error {
	return nil
}

func (n *NoopJournalTracker) LastCursor() (string, error) {
	return "", nil
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
