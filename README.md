# CachedPath - Go

A Go library inspired by the Python [cached_path](https://github.com/allenai/cached_path), providing a unified and simple interface to access local and remote files with automatic caching.

## Features

- ✅ **Unified Interface**: Same function for local files and remote URLs
- 🌐 **Multiple Protocols**: Support for HTTP, HTTPS (S3 and GCS can be added)
- 💾 **Automatic Cache**: Downloads are stored locally in cache
- 🔒 **Thread-Safe**: File locking prevents race conditions in concurrent downloads
- 📦 **Archive Extraction**: Automatic support for ZIP and TAR.GZ
- 📊 **Progress Bar**: Visual feedback during downloads
- 🔑 **Custom Headers**: Support for authentication and custom HTTP headers
- 🎯 **Special Syntax**: Access specific files inside archives with `!`
- ⚙️ **Functional Options**: Idiomatic Go pattern for flexible configuration
- 🔄 **Automatic Retry**: Automatic retries on failure
- ⏱️ **Configurable Timeout**: Full control over HTTP timeouts
- 🛠️ **Custom HTTP Client**: Use your own `http.Client`

## Installation

```bash
go get github.com/CezarGarrido/cachedpath
```

## Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/CezarGarrido/cachedpath"
)

func main() {
    // Simple download
    path, err := cachedpath.CachedPath("https://example.com/model.bin")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Cached file:", path)
}
```

## Examples with Functional Options

### 1. Local File

```go
path, err := cachedpath.CachedPath("/path/to/local/file.txt")
```

### 2. Download with Custom Timeout

```go
path, err := cachedpath.CachedPath(
    "https://example.com/large-file.bin",
    cachedpath.WithTimeout(120 * time.Second),
)
```

### 3. Authentication Headers

```go
// Bearer token
path, err := cachedpath.CachedPath(
    "https://api.example.com/private/file.bin",
    cachedpath.WithAuth("YOUR_TOKEN"),
)

// Custom headers
path, err := cachedpath.CachedPath(
    "https://api.example.com/file.bin",
    cachedpath.WithHeader("Authorization", "Bearer TOKEN"),
    cachedpath.WithUserAgent("MyApp/1.0"),
)
```

### 4. Retry Configuration

```go
path, err := cachedpath.CachedPath(
    "https://unstable-server.com/file.bin",
    cachedpath.WithMaxRetries(5),
    cachedpath.WithRetryDelay(2 * time.Second),
)
```

### 5. Automatic Archive Extraction

```go
// Extracts entire archive
dirPath, err := cachedpath.CachedPath(
    "https://example.com/archive.tar.gz",
    cachedpath.WithExtractArchive(true),
)

// Access specific file inside archive
path, err := cachedpath.CachedPath(
    "https://example.com/model.tar.gz!weights/model.bin",
)
```

### 6. Custom Cache Directory

```go
path, err := cachedpath.CachedPath(
    "https://example.com/file.bin",
    cachedpath.WithCacheDir("/custom/cache/dir"),
)
```

### 7. Quiet Mode

```go
path, err := cachedpath.CachedPath(
    "https://example.com/file.bin",
    cachedpath.WithQuiet(true),
)
```

### 8. Custom HTTP Client

```go
customClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        50,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    },
}

path, err := cachedpath.CachedPath(
    "https://example.com/file.bin",
    cachedpath.WithHTTPClient(customClient),
)
```

### 9. Combining Multiple Options

```go
path, err := cachedpath.CachedPath(
    "https://api.example.com/data.tar.gz",
    cachedpath.WithCacheDir("/tmp/my_cache"),
    cachedpath.WithTimeout(120 * time.Second),
    cachedpath.WithMaxRetries(5),
    cachedpath.WithRetryDelay(2 * time.Second),
    cachedpath.WithAuth("my-secret-token"),
    cachedpath.WithUserAgent("DataProcessor/2.0"),
    cachedpath.WithExtractArchive(true),
    cachedpath.WithQuiet(false),
)
```

## Available Options

### Functional Options

| Function | Description | Default |
|----------|-------------|---------|
| `WithCacheDir(dir)` | Sets cache directory | `~/.cache/cached_path/` |
| `WithExtractArchive(bool)` | Automatically extracts archives | `false` |
| `WithForceExtract(bool)` | Forces extraction even if already exists | `false` |
| `WithQuiet(bool)` | Suppresses progress messages | `false` |
| `WithProgress(display)` | Sets custom progress display | `nil` |
| `WithHeaders(map)` | Sets custom HTTP headers | `{}` |
| `WithHeader(key, value)` | Adds an HTTP header | - |
| `WithHTTPClient(client)` | Sets custom HTTP client | Default client |
| `WithTimeout(duration)` | Sets timeout for requests | `30s` |
| `WithMaxRetries(n)` | Sets maximum retry attempts | `3` |
| `WithRetryDelay(duration)` | Sets delay between retries | `1s` |
| `WithAuth(token)` | Adds Bearer token | - |
| `WithUserAgent(ua)` | Sets User-Agent | `CachedPath-Go/1.0` |

## Cache Directory Configuration

The cache directory can be configured in three ways (in order of priority):

1. **`WithCacheDir` option**:
   ```go
   path, err := cachedpath.CachedPath(
       url,
       cachedpath.WithCacheDir("/custom/cache/dir"),
   )
   ```

2. **Environment variable `CACHED_PATH_CACHE_ROOT`**:
   ```bash
   export CACHED_PATH_CACHE_ROOT=/custom/cache/dir
   ```

3. **Default**: `~/.cache/cached_path/`

## Supported Protocols

- ✅ `http://` - HTTP
- ✅ `https://` - HTTPS
- 🔜 `s3://` - AWS S3 (planned)
- 🔜 `gs://` - Google Cloud Storage (planned)

## Supported Archive Formats

- ✅ `.zip` - ZIP
- ✅ `.tar.gz` - TAR with GZIP
- ✅ `.tgz` - TAR with GZIP (abbreviated)

## Architecture

```
cachedpath/
├── cachedpath.go      # Main CachedPath() function
├── options.go         # Functional Options
├── archive.go         # Archive extraction
├── schemes/
│   ├── scheme.go      # SchemeClient interface
│   ├── http.go        # HTTP/HTTPS client with retry
│   └── ...            # Other clients
├── filelock.go        # File locking system
├── meta.go            # Cache metadata
├── progress.go        # Progress bar
├── util.go            # Utility functions
└── errors.go          # Custom errors
```

## Advanced Features

### Automatic Retry

The library implements automatic retry for HTTP requests that fail due to:
- Temporary network errors
- Timeouts
- Server 5xx errors

```go
path, err := cachedpath.CachedPath(
    url,
    cachedpath.WithMaxRetries(5),        // Retry up to 5 times
    cachedpath.WithRetryDelay(2*time.Second), // Wait 2s between retries
)
```

The delay between retries increases progressively (linear backoff).

### Custom HTTP Client

You can provide your own `http.Client` for full control:

```go
client := &http.Client{
    Timeout: 120 * time.Second,
    Transport: &http.Transport{
        Proxy: http.ProxyFromEnvironment,
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   10,
        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:   10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    },
}

path, err := cachedpath.CachedPath(url, cachedpath.WithHTTPClient(client))
```

### Thread Safety

The library is thread-safe and uses file locking to prevent race conditions when multiple processes or goroutines try to download the same file simultaneously.

```go
// Safe to use in concurrent goroutines
var wg sync.WaitGroup
urls := []string{"url1", "url2", "url3"}

for _, url := range urls {
    wg.Add(1)
    go func(u string) {
        defer wg.Done()
        path, err := cachedpath.CachedPath(u)
        // ...
    }(url)
}
wg.Wait()
```

## Testing

Run tests with:

```bash
# Unit tests (fast)
go test -short ./tests/

# All tests including integration (requires internet)
go test -v ./tests/

# With coverage
go test -cover ./tests/
```

## Comparison with Python Version

| Feature | Python | Go |
|---------|--------|-----|
| HTTP/HTTPS | ✅ | ✅ |
| AWS S3 | ✅ | 🔜 |
| Google Cloud Storage | ✅ | 🔜 |
| HuggingFace Hub | ✅ | 🔜 |
| Archive Extraction | ✅ | ✅ |
| File Locking | ✅ | ✅ |
| Progress Bar | ✅ | ✅ |
| Custom Headers | ✅ | ✅ |
| Syntax with `!` | ✅ | ✅ |
| Automatic Retry | ❌ | ✅ |
| Configurable Timeout | ⚠️ Limited | ✅ |
| Functional Options | ❌ | ✅ |

## Advantages of Functional Options Pattern

1. **Clean Code**: More readable function calls
2. **Backward Compatibility**: Adding new options doesn't break existing code
3. **Default Values**: Sensible options without configuration
4. **Composition**: Combine options flexibly
5. **Type-Safe**: Errors detected at compile time

## Complete Examples

See the `examples/` directory for complete usage examples:

```bash
cd examples
go run main.go
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a branch for your feature (`git checkout -b feature/new-feature`)
3. Commit your changes (`git commit -am 'Add new feature'`)
4. Push to the branch (`git push origin feature/new-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details.

## Inspiration

This project was inspired by the excellent [cached_path](https://github.com/allenai/cached_path) library from Allen Institute for AI (AllenAI).

## Roadmap

### v1.1 (Next version)
- [ ] AWS S3 support
- [ ] Google Cloud Storage support
- [ ] HuggingFace Hub support
- [ ] Progress bar improvements

### v1.2
- [ ] Cache with TTL (Time To Live)
- [ ] Cache compression
- [ ] Cache usage metrics
- [ ] Support for resuming interrupted downloads

### v2.0
- [ ] Streaming API
- [ ] Support for multiple cache backends
- [ ] Plugin system for new protocols

## Changelog

### v1.1.0 (Current)
- ✅ Implemented Functional Options pattern
- ✅ Added automatic retry with backoff
- ✅ Configurable timeout
- ✅ Custom HTTP client
- ✅ Documentation improvements

### v1.0.0
- ✅ Initial implementation
- ✅ HTTP/HTTPS support
- ✅ Automatic caching
- ✅ Archive extraction
- ✅ File locking# cachedpath
