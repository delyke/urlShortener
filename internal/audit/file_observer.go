package audit

import (
	"context"
	"encoding/json"
	"os"
)

// FileObserver writes audit events to a local file
type FileObserver struct {
	path string
}

// NewFileObserver creates an observer that appends audit events to a file
func NewFileObserver(path string) (*FileObserver, error) {
	if path == "" {
		return nil, ErrAuditFileIsEmpty
	}
	return &FileObserver{path: path}, nil
}

// OnEvent serializes the event as JSON and appends it to the file.
func (o *FileObserver) OnEvent(_ context.Context, e Event) error {
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(o.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(payload, '\n'))
	return err
}
