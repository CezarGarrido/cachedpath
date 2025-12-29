package cachedpath

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/CezarGarrido/cachedpath/schemes"
)

// CachedPath is the main function that determines if the input is a URL or local path.
// If it's a URL, it downloads the file and stores it in cache, returning the cached file path.
// If it's a local path, it checks if the file exists and returns the path.
//
// Basic example:
//
//	path, err := cachedpath.CachedPath("https://example.com/file.bin")
//
// Example with options:
//
//	path, err := cachedpath.CachedPath(
//	    "https://example.com/file.tar.gz",
//	    cachedpath.WithExtractArchive(true),
//	    cachedpath.WithTimeout(60 * time.Second),
//	)
func CachedPath(urlOrFilename string, opts ...Option) (string, error) {
	// Apply default options
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Ensure cache directory exists
	if err := EnsureDir(options.CacheDir); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Check for special archive syntax (file.tar.gz!internal/path)
	archivePath, internalPath, hasInternalPath := ParseArchivePath(urlOrFilename)

	// Determine if it's a URL or local path
	if !IsURL(archivePath) {
		// It's a local path
		return handleLocalPath(archivePath, internalPath, hasInternalPath, options)
	}

	// It's a remote URL
	return handleRemoteURL(archivePath, internalPath, hasInternalPath, options)
}

// handleLocalPath processes local paths
func handleLocalPath(path, internalPath string, hasInternalPath bool, opts *Options) (string, error) {
	// Check if file exists
	if !FileExists(path) {
		return "", fmt.Errorf("%w: %s", ErrFileNotFound, path)
	}

	// If there's an internal path, extract the specific file from the archive
	if hasInternalPath {
		if !IsArchive(path) {
			return "", fmt.Errorf("file is not an archive: %s", path)
		}

		extractDir := filepath.Join(opts.CacheDir, "extracted", filepath.Base(path))
		extractedPath, err := ExtractSpecificFile(path, internalPath, extractDir)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrExtractionFailed, err)
		}
		return extractedPath, nil
	}

	// If should extract archive
	if opts.ExtractArchive && IsArchive(path) {
		extractDir := filepath.Join(opts.CacheDir, "extracted", filepath.Base(path))

		// Check if already extracted
		if !opts.ForceExtract && FileExists(extractDir) {
			return extractDir, nil
		}

		if err := ExtractArchive(path, extractDir); err != nil {
			return "", fmt.Errorf("%w: %v", ErrExtractionFailed, err)
		}
		return extractDir, nil
	}

	return path, nil
}

// handleRemoteURL processes remote URLs
func handleRemoteURL(url, internalPath string, hasInternalPath bool, opts *Options) (string, error) {
	// Get URL scheme
	scheme := GetScheme(url)
	if scheme == "" {
		return "", ErrInvalidURL
	}

	// Normalize scheme (https also uses http client)
	if scheme == "https" {
		scheme = "http"
	}

	// Get appropriate client
	client, ok := schemes.GetClient(scheme)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedScheme, scheme)
	}

	// Configure HTTP client if it's HTTPClient
	if httpClient, ok := client.(*schemes.HTTPClient); ok {
		httpClient.SetHTTPClient(opts.getHTTPClient())
		httpClient.SetRetryConfig(opts.MaxRetries, opts.RetryDelay)
	}

	// Get ETag for versioning
	etag, err := client.GetETag(url, opts.Headers)
	if err != nil {
		// If fails to get ETag, continue without it
		etag = ""
	}

	// Generate cache filename
	filename := ResourceToFilename(url, etag)
	cachePath := filepath.Join(opts.CacheDir, filename)

	// Use file lock to prevent concurrent downloads
	lockPath := LockFilePath(cachePath)

	err = WithLock(lockPath, func() error {
		// Check if already in cache
		if FileExists(cachePath) {
			// Check metadata
			metaPath := MetaFilePath(cachePath)
			if FileExists(metaPath) {
				meta, err := LoadMetaFromFile(metaPath)
				if err == nil && meta.ETag == etag {
					// Cache is up to date
					return nil
				}
			}
		}

		// Download the file
		return downloadFile(client, url, cachePath, opts)
	})

	if err != nil {
		return "", err
	}

	// Save metadata
	meta := NewMeta(url, cachePath, etag)
	metaPath := MetaFilePath(cachePath)
	if err := meta.SaveToFile(metaPath); err != nil {
		// Not critical if fails to save metadata
		fmt.Fprintf(os.Stderr, "Warning: failed to save metadata: %v\n", err)
	}

	// If there's an internal path, extract the specific file
	if hasInternalPath {
		if !IsArchive(cachePath) {
			return "", fmt.Errorf("file is not an archive: %s", cachePath)
		}

		extractDir := filepath.Join(opts.CacheDir, "extracted", filename)
		extractedPath, err := ExtractSpecificFile(cachePath, internalPath, extractDir)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrExtractionFailed, err)
		}
		return extractedPath, nil
	}

	// If should extract archive
	if opts.ExtractArchive && IsArchive(cachePath) {
		extractDir := filepath.Join(opts.CacheDir, "extracted", filename)

		// Check if already extracted
		if !opts.ForceExtract && FileExists(extractDir) {
			return extractDir, nil
		}

		if err := ExtractArchive(cachePath, extractDir); err != nil {
			return "", fmt.Errorf("%w: %v", ErrExtractionFailed, err)
		}
		return extractDir, nil
	}

	return cachePath, nil
}

// downloadFile downloads a file using the appropriate client
func downloadFile(client schemes.SchemeClient, url, destPath string, opts *Options) error {
	// Get file size
	size, err := client.GetSize(url, opts.Headers)
	if err != nil {
		size = 0 // Continue without size
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp(filepath.Dir(destPath), ".download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Remove on error

	// Configure progress
	progress := opts.Progress
	if progress == nil {
		progress = NewSimpleProgress(opts.Quiet)
	}

	progress.Start(size, url)
	defer progress.Finish()

	// Create writer with progress
	writer := NewProgressWriter(tmpFile, progress)

	// Download the file
	err = client.GetResource(url, writer, opts.Headers)
	tmpFile.Close()

	if err != nil {
		return fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	// Move temporary file to final destination
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to move downloaded file: %w", err)
	}

	return nil
}
