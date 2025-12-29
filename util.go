package cachedpath

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// IsURL checks if a string is a valid URL
func IsURL(path string) bool {
	u, err := url.Parse(path)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// GetScheme extracts the scheme from a URL
func GetScheme(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Scheme
}

// ResourceToFilename converts a URL and ETag into a unique filename
func ResourceToFilename(resourceURL, etag string) string {
	// Create a hash of URL + ETag to generate unique name
	hash := sha256.New()
	hash.Write([]byte(resourceURL))
	if etag != "" {
		hash.Write([]byte(etag))
	}
	hashStr := hex.EncodeToString(hash.Sum(nil))

	// Extract extension from URL if possible
	u, _ := url.Parse(resourceURL)
	ext := filepath.Ext(u.Path)

	if ext != "" {
		return hashStr + ext
	}
	return hashStr
}

// ParseArchivePath parses paths in the format "file.tar.gz!internal/path"
func ParseArchivePath(path string) (archivePath, internalPath string, ok bool) {
	parts := strings.SplitN(path, "!", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return path, "", false
}

// GetDefaultCacheDir returns the default cache directory
func GetDefaultCacheDir() (string, error) {
	// Check environment variable
	if dir := os.Getenv("CACHED_PATH_CACHE_ROOT"); dir != "" {
		return dir, nil
	}

	// Use user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".cache", "cached_path"), nil
}

// EnsureDir ensures a directory exists
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// LockFilePath returns the lock file path
func LockFilePath(cachePath string) string {
	return cachePath + ".lock"
}

// MetaFilePath returns the metadata file path
func MetaFilePath(cachePath string) string {
	return cachePath + ".meta.json"
}
