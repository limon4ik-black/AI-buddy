package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
)

type UpstreamError struct{ StatusCode int }

func (e *UpstreamError) Error() string { return fmt.Sprintf("upstream status %d", e.StatusCode) }

func Call[T any](ctx context.Context, client *http.Client, method, url, secret, bearer string, body any) (T, error) {
	var result T
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return result, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(data))
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", secret)
	req.Header.Set("Authorization", bearer)
	if id := httpapi.RequestID(ctx); id != "" {
		req.Header.Set("X-Request-ID", id)
	}
	res, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return result, &UpstreamError{StatusCode: res.StatusCode}
	}
	err = json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(&result)
	return result, err
}
