package audit

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Log writes newline-delimited JSON data to a file with simple size-based rotation.
type Log struct {
	path    string
	maxSize int64

	mu   sync.Mutex
	file *os.File
	size int64
}

// Open creates or appends to the log file at path. max specifies the
// maximum size in bytes before the file is rotated. If max is 0, no rotation
// occurs.
func Open(path string, max int64) (*Log, error) {
	if path == "" {
		return nil, fmt.Errorf("path required")
	}
	l := &Log{path: path, maxSize: max}
	if err := l.open(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Log) open() error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	l.file = f
	fi, err := f.Stat()
	if err == nil {
		l.size = fi.Size()
	}
	return nil
}

func (l *Log) rotate() error {
	if l.maxSize <= 0 || l.size < l.maxSize {
		return nil
	}
	if l.file != nil {
		l.file.Close()
	}
	ts := time.Now().UTC().Format("20060102150405")
	newName := fmt.Sprintf("%s.%s", l.path, ts)
	if err := os.Rename(l.path, newName); err != nil {
		return err
	}
	return l.open()
}

// Write implements io.Writer.
func (l *Log) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.rotate(); err != nil {
		return 0, err
	}
	n, err := l.file.Write(p)
	l.size += int64(n)
	return n, err
}

// Close closes the underlying file.
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}

// Path returns the current log file path.
func (l *Log) Path() string { return l.path }
