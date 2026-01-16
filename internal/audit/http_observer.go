package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPObserver struct {
	url    string
	client *http.Client
}

func NewHTTPObserver(url string, client *http.Client) (*HTTPObserver, error) {
	if url == "" {
		return nil, ErrAuditUrlIsEmpty
	}
	if client == nil {
		client = &http.Client{Timeout: time.Second * 5}
	}
	return &HTTPObserver{url, client}, nil
}

func (o *HTTPObserver) OnEvent(ctx context.Context, e Event) error {
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("эндпоинт для аудита вернул статус %d", resp.StatusCode)
	}
	return nil
}
