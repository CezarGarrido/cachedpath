package schemes

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// HTTPClient implementa SchemeClient para HTTP e HTTPS
type HTTPClient struct {
	client     *http.Client
	maxRetries int
	retryDelay time.Duration
}

// NewHTTPClient creates a new HTTPClient with default settings
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}
}

// SetHTTPClient define um cliente HTTP customizado
func (c *HTTPClient) SetHTTPClient(client *http.Client) {
	if client != nil {
		c.client = client
	}
}

// SetRetryConfig sets the retry configuration
func (c *HTTPClient) SetRetryConfig(maxRetries int, retryDelay time.Duration) {
	c.maxRetries = maxRetries
	c.retryDelay = retryDelay
}

// doRequestWithRetry executes a request with automatic retry
func (c *HTTPClient) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying
			time.Sleep(c.retryDelay * time.Duration(attempt))
		}

		resp, err = c.client.Do(req)

		// Sucesso
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		// If not a network error or timeout, don't retry
		if err == nil {
			// Status code different from 200
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				// 4xx errors are generally not recoverable
				return resp, nil
			}
		}

		// Fecha response anterior se houver
		if resp != nil {
			resp.Body.Close()
		}

		// If it's the last attempt, return the error
		if attempt == c.maxRetries {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed after %d retries: %w", c.maxRetries, err)
	}

	return resp, nil
}

// GetResource baixa o recurso via HTTP/HTTPS
func (c *HTTPClient) GetResource(url string, writer io.Writer, headers map[string]string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add default User-Agent if not provided
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "CachedPath-Go/1.0")
	}

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d %s", resp.StatusCode, resp.Status)
	}

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

// GetSize retorna o tamanho do recurso
func (c *HTTPClient) GetSize(url string, headers map[string]string) (int64, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add default User-Agent if not provided
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "CachedPath-Go/1.0")
	}

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get size: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HEAD request failed with status: %d %s", resp.StatusCode, resp.Status)
	}

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, nil
	}

	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse content length: %w", err)
	}

	return size, nil
}

// GetETag retorna o ETag do recurso
func (c *HTTPClient) GetETag(url string, headers map[string]string) (string, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add default User-Agent if not provided
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "CachedPath-Go/1.0")
	}

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return "", fmt.Errorf("failed to get ETag: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HEAD request failed with status: %d %s", resp.StatusCode, resp.Status)
	}

	etag := resp.Header.Get("ETag")
	if etag == "" {
		// If no ETag, use Last-Modified as alternative
		etag = resp.Header.Get("Last-Modified")
	}

	return etag, nil
}

// Scheme retorna o nome do esquema
func (c *HTTPClient) Scheme() string {
	return "http" // Funciona para http e https
}
