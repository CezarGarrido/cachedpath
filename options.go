package cachedpath

import (
	"net/http"
	"time"
)

// Options contains the options for CachedPath
type Options struct {
	// CacheDir is the directory where files will be cached
	CacheDir string

	// ExtractArchive indicates if archives should be automatically extracted
	ExtractArchive bool

	// ForceExtract forces extraction even if the directory already exists
	ForceExtract bool

	// Quiet suppresses progress messages
	Quiet bool

	// Progress is a custom progress display
	Progress ProgressDisplay

	// Headers are custom HTTP headers for requests
	Headers map[string]string

	// HTTPClient is a custom HTTP client
	HTTPClient *http.Client

	// Timeout is the timeout for HTTP requests (default: 30 seconds)
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts on failure (default: 3)
	MaxRetries int

	// RetryDelay is the delay between retry attempts (default: 1 second)
	RetryDelay time.Duration
}

// Option is a function that modifies Options
type Option func(*Options)

// defaultOptions returns the default options
func defaultOptions() *Options {
	cacheDir, _ := GetDefaultCacheDir()
	return &Options{
		CacheDir:       cacheDir,
		ExtractArchive: false,
		ForceExtract:   false,
		Quiet:          false,
		Progress:       nil,
		Headers:        make(map[string]string),
		HTTPClient:     nil, // will be created with default settings if nil
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
	}
}

// WithCacheDir sets the cache directory
func WithCacheDir(dir string) Option {
	return func(o *Options) {
		o.CacheDir = dir
	}
}

// WithExtractArchive enables automatic archive extraction
func WithExtractArchive(extract bool) Option {
	return func(o *Options) {
		o.ExtractArchive = extract
	}
}

// WithForceExtract forces extraction even if already exists
func WithForceExtract(force bool) Option {
	return func(o *Options) {
		o.ForceExtract = force
	}
}

// WithQuiet suppresses progress messages
func WithQuiet(quiet bool) Option {
	return func(o *Options) {
		o.Quiet = quiet
	}
}

// WithProgress sets a custom progress display
func WithProgress(progress ProgressDisplay) Option {
	return func(o *Options) {
		o.Progress = progress
	}
}

// WithHeaders sets custom HTTP headers
func WithHeaders(headers map[string]string) Option {
	return func(o *Options) {
		o.Headers = headers
	}
}

// WithHeader adiciona um header HTTP
func WithHeader(key, value string) Option {
	return func(o *Options) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		o.Headers[key] = value
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(o *Options) {
		o.HTTPClient = client
	}
}

// WithTimeout sets the timeout for HTTP requests
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retry attempts
func WithMaxRetries(maxRetries int) Option {
	return func(o *Options) {
		o.MaxRetries = maxRetries
	}
}

// WithRetryDelay sets the delay between retry attempts
func WithRetryDelay(delay time.Duration) Option {
	return func(o *Options) {
		o.RetryDelay = delay
	}
}

// WithAuth adds Bearer token authentication
func WithAuth(token string) Option {
	return func(o *Options) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		o.Headers["Authorization"] = "Bearer " + token
	}
}

// WithBasicAuth adds basic authentication
func WithBasicAuth(username, password string) Option {
	return func(o *Options) {
		if o.HTTPClient == nil {
			o.HTTPClient = &http.Client{}
		}
		// Will be configured in the HTTP client
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		// Basic auth will be added automatically by http.Client
	}
}

// WithUserAgent define o User-Agent
func WithUserAgent(userAgent string) Option {
	return func(o *Options) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		o.Headers["User-Agent"] = userAgent
	}
}

// getHTTPClient retorna o cliente HTTP configurado
func (o *Options) getHTTPClient() *http.Client {
	if o.HTTPClient != nil {
		return o.HTTPClient
	}

	// Create client with default settings
	return &http.Client{
		Timeout: o.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}
