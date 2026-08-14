package glances

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "http://192.168.0.44:61208"

const fsEndpoint = "/api/4/fs"

// Client communicates with the Glances REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// NewClient returns a new Client for the Glances server.
// By default it points to the server defined in this package.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithBaseURL overrides the Glances server base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// WithHTTPClient sets the HTTP client used to make requests.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout configures a timeout for all requests made by the client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// FileSystem represents the storage usage of a mounted filesystem.
type FileSystem struct {
	DeviceName string  `json:"device_name"`
	FSType     string  `json:"fs_type"`
	MountPoint string  `json:"mnt_point"`
	Options    string  `json:"options"`
	Size       int64   `json:"size"`
	Used       int64   `json:"used"`
	Free       int64   `json:"free"`
	Percent    float64 `json:"percent"`
	Key        string  `json:"key"`
}

// GetFS returns the storage usage of every filesystem reported by Glances.
func (c *Client) GetFS(ctx context.Context) ([]FileSystem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fsEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch filesystems: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("glances API returned status: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var filesystems []FileSystem
	if err := json.Unmarshal(body, &filesystems); err != nil {
		return nil, fmt.Errorf("failed to decode filesystems: %w", err)
	}
	return filesystems, nil
}
