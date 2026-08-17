package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"time"
)

type Client struct {
	httpClient *nethttp.Client
}

const defaultTimeout = 5 * time.Second

func New() *Client {
	return &Client{
		httpClient: &nethttp.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (c *Client) DoJSON(ctx context.Context, method, url string, out any) error {
	req, err := nethttp.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       body,
		}
	}

	if out == nil {
		return nil
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: %s", e.Status)
}

func (e *HTTPError) IsNotFound() bool {
	return e.StatusCode == nethttp.StatusNotFound
}

func (e *HTTPError) IsRateLimited() bool {
	return e.StatusCode == nethttp.StatusTooManyRequests
}

func (e *HTTPError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

func (e *HTTPError) IsServerError() bool {
	return e.StatusCode >= 500
}
