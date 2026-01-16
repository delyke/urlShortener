package audit

import (
	"context"
	"encoding/json"
	"os"
)

type FileObserver struct {
	path string
}

func NewFileObserver(path string) (*FileObserver, error) {
	if path == "" {
		return nil, ErrAuditFileIsEmpty
	}
	return &FileObserver{path: path}, nil
}

func (o *FileObserver) OnEvent(ctx context.Context, e Event) error {
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
