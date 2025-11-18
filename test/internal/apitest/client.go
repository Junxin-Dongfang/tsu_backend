package apitest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// APIResponse matches the project's generic API envelope.
type APIResponse[T any] struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      *T     `json:"data"`
	Timestamp int64  `json:"timestamp"`
	TraceID   string `json:"trace_id"`
}

// Client is a lightweight HTTP helper for API tests.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Status safely returns HTTP status code (0 if resp is nil).
func Status(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

// NewClient creates a client with sane defaults.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// PostJSON sends a JSON POST.
func PostJSON[T any, R any](ctx context.Context, c *Client, path string, payload T, token string) (*APIResponse[R], *http.Response, []byte, error) {
	return doJSON[T, R](ctx, c, http.MethodPost, path, &payload, token)
}

// PutJSON sends a JSON PUT.
func PutJSON[T any, R any](ctx context.Context, c *Client, path string, payload T, token string) (*APIResponse[R], *http.Response, []byte, error) {
	return doJSON[T, R](ctx, c, http.MethodPut, path, &payload, token)
}

// GetJSON sends a JSON GET.
func GetJSON[R any](ctx context.Context, c *Client, path string, token string) (*APIResponse[R], *http.Response, []byte, error) {
	return doJSON[struct{}, R](ctx, c, http.MethodGet, path, nil, token)
}

// DeleteJSON sends a JSON DELETE with optional payload.
func DeleteJSON[T any, R any](ctx context.Context, c *Client, path string, payload *T, token string) (*APIResponse[R], *http.Response, []byte, error) {
	return doJSON[T, R](ctx, c, http.MethodDelete, path, payload, token)
}

func doJSON[T any, R any](ctx context.Context, c *Client, method, path string, payload *T, token string) (*APIResponse[R], *http.Response, []byte, error) {
	var body io.Reader
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(buf)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Session-Token", token)
		req.Header.Set("Cookie", "ory_kratos_session="+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, resp, nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, nil, fmt.Errorf("read response: %w", err)
	}

	var apiResp APIResponse[R]
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, resp, bodyBytes, fmt.Errorf("decode response: %w", err)
	}

	return &apiResp, resp, bodyBytes, nil
}
