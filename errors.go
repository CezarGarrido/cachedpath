package cachedpath

import "errors"

var (
	// ErrInvalidURL indicates that the provided URL is invalid
	ErrInvalidURL = errors.New("invalid URL")

	// ErrFileNotFound indicates that the file was not found
	ErrFileNotFound = errors.New("file not found")

	// ErrUnsupportedScheme indicates that the URL scheme is not supported
	ErrUnsupportedScheme = errors.New("unsupported URL scheme")

	// ErrDownloadFailed indicates that the download failed
	ErrDownloadFailed = errors.New("download failed")

	// ErrExtractionFailed indicates that file extraction failed
	ErrExtractionFailed = errors.New("extraction failed")

	// ErrLockFailed indicates that it was not possible to acquire the file lock
	ErrLockFailed = errors.New("failed to acquire file lock")
)
