package tests

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CezarGarrido/cachedpath"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com/file.txt", true},
		{"http://example.com", true},
		{"s3://bucket/key", true},
		{"/local/path/file.txt", false},
		{"./relative/path", false},
		{"file.txt", false},
	}

	for _, tt := range tests {
		result := cachedpath.IsURL(tt.input)
		if result != tt.expected {
			t.Errorf("IsURL(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestGetScheme(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com/file.txt", "https"},
		{"http://example.com", "http"},
		{"s3://bucket/key", "s3"},
		{"gs://bucket/object", "gs"},
		{"/local/path", ""},
	}

	for _, tt := range tests {
		result := cachedpath.GetScheme(tt.input)
		if result != tt.expected {
			t.Errorf("GetScheme(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseArchivePath(t *testing.T) {
	tests := []struct {
		input        string
		archivePath  string
		internalPath string
		ok           bool
	}{
		{"model.tar.gz!weights.th", "model.tar.gz", "weights.th", true},
		{"archive.zip!data/file.txt", "archive.zip", "data/file.txt", true},
		{"regular_file.txt", "regular_file.txt", "", false},
	}

	for _, tt := range tests {
		archive, internal, ok := cachedpath.ParseArchivePath(tt.input)
		if archive != tt.archivePath || internal != tt.internalPath || ok != tt.ok {
			t.Errorf("ParseArchivePath(%q) = (%q, %q, %v), expected (%q, %q, %v)",
				tt.input, archive, internal, ok, tt.archivePath, tt.internalPath, tt.ok)
		}
	}
}

func TestResourceToFilename(t *testing.T) {
	url1 := "https://example.com/file.txt"
	etag1 := "abc123"

	filename1 := cachedpath.ResourceToFilename(url1, etag1)
	if filename1 == "" {
		t.Error("ResourceToFilename returned empty string")
	}

	// Mesmo URL e ETag devem gerar mesmo filename
	filename2 := cachedpath.ResourceToFilename(url1, etag1)
	if filename1 != filename2 {
		t.Error("ResourceToFilename not deterministic")
	}

	// URL diferente deve gerar filename diferente
	url2 := "https://example.com/other.txt"
	filename3 := cachedpath.ResourceToFilename(url2, etag1)
	if filename1 == filename3 {
		t.Error("Different URLs generated same filename")
	}
}

func TestIsArchive(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"file.zip", true},
		{"file.tar.gz", true},
		{"file.tgz", true},
		{"file.txt", false},
		{"file.pdf", false},
	}

	for _, tt := range tests {
		result := cachedpath.IsArchive(tt.path)
		if result != tt.expected {
			t.Errorf("IsArchive(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}

func TestCachedPathLocalFile(t *testing.T) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("test content")
	tmpFile.Close()

	// Testa CachedPath com arquivo local
	path, err := cachedpath.CachedPath(tmpFile.Name())
	if err != nil {
		t.Errorf("CachedPath failed for local file: %v", err)
	}

	if path != tmpFile.Name() {
		t.Errorf("CachedPath returned wrong path: got %q, expected %q", path, tmpFile.Name())
	}
}

func TestCachedPathNonExistentFile(t *testing.T) {
	// Test with non-existent file
	_, err := cachedpath.CachedPath("/non/existent/file.txt")
	if err == nil {
		t.Error("CachedPath should fail for non-existent file")
	}
}

func TestCachedPathWithOptions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cachedpath-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test with WithCacheDir option
	path, err := cachedpath.CachedPath(
		tmpFile.Name(),
		cachedpath.WithCacheDir(tmpDir),
	)
	if err != nil {
		t.Errorf("CachedPath with options failed: %v", err)
	}

	if path != tmpFile.Name() {
		t.Errorf("CachedPath returned wrong path: %s", path)
	}
}

func TestCachedPathHTTPS(t *testing.T) {
	// Integration test - requires internet connection
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "cachedpath-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Baixa um arquivo pequeno
	url := "https://raw.githubusercontent.com/golang/go/master/LICENSE"
	path, err := cachedpath.CachedPath(
		url,
		cachedpath.WithCacheDir(tmpDir),
		cachedpath.WithQuiet(true),
	)
	if err != nil {
		t.Errorf("CachedPath failed for HTTPS URL: %v", err)
	}

	// Verifica se o arquivo foi criado
	if !cachedpath.FileExists(path) {
		t.Errorf("Downloaded file does not exist: %s", path)
	}

	// Check if it's in the correct cache directory
	if !filepath.HasPrefix(path, tmpDir) {
		t.Errorf("Downloaded file not in cache dir: %s", path)
	}

	// Segunda chamada deve usar cache
	path2, err := cachedpath.CachedPath(
		url,
		cachedpath.WithCacheDir(tmpDir),
		cachedpath.WithQuiet(true),
	)
	if err != nil {
		t.Errorf("Second CachedPath call failed: %v", err)
	}

	if path != path2 {
		t.Errorf("Second call returned different path: %s vs %s", path, path2)
	}
}

func TestWithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "cachedpath-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Testa com timeout muito curto (deve falhar)
	_, err = cachedpath.CachedPath(
		"https://httpbin.org/delay/5",
		cachedpath.WithCacheDir(tmpDir),
		cachedpath.WithTimeout(1*time.Second),
		cachedpath.WithQuiet(true),
	)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestWithCustomHTTPClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "cachedpath-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	customClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	path, err := cachedpath.CachedPath(
		"https://golang.org/robots.txt",
		cachedpath.WithCacheDir(tmpDir),
		cachedpath.WithHTTPClient(customClient),
		cachedpath.WithQuiet(true),
	)
	if err != nil {
		t.Errorf("CachedPath with custom client failed: %v", err)
	}

	if !cachedpath.FileExists(path) {
		t.Errorf("Downloaded file does not exist: %s", path)
	}
}

func TestWithHeaders(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "cachedpath-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	path, err := cachedpath.CachedPath(
		"https://httpbin.org/headers",
		cachedpath.WithCacheDir(tmpDir),
		cachedpath.WithUserAgent("TestAgent/1.0"),
		cachedpath.WithHeader("X-Test-Header", "test-value"),
		cachedpath.WithQuiet(true),
	)
	if err != nil {
		t.Errorf("CachedPath with headers failed: %v", err)
	}

	if !cachedpath.FileExists(path) {
		t.Errorf("Downloaded file does not exist: %s", path)
	}
}

func TestGetDefaultCacheDir(t *testing.T) {
	// Save original value
	originalEnv := os.Getenv("CACHED_PATH_CACHE_ROOT")
	defer os.Setenv("CACHED_PATH_CACHE_ROOT", originalEnv)

	// Test with environment variable
	testDir := "/tmp/test_cache"
	os.Setenv("CACHED_PATH_CACHE_ROOT", testDir)

	cacheDir, err := cachedpath.GetDefaultCacheDir()
	if err != nil {
		t.Errorf("GetDefaultCacheDir failed: %v", err)
	}

	if cacheDir != testDir {
		t.Errorf("GetDefaultCacheDir returned %q, expected %q", cacheDir, testDir)
	}

	// Test without environment variable
	os.Unsetenv("CACHED_PATH_CACHE_ROOT")

	cacheDir, err = cachedpath.GetDefaultCacheDir()
	if err != nil {
		t.Errorf("GetDefaultCacheDir failed: %v", err)
	}

	if cacheDir == "" {
		t.Error("GetDefaultCacheDir returned empty string")
	}
}

func TestMultipleOptions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cachedpath-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test with multiple options
	path, err := cachedpath.CachedPath(
		tmpFile.Name(),
		cachedpath.WithCacheDir(tmpDir),
		cachedpath.WithQuiet(true),
		cachedpath.WithTimeout(30*time.Second),
		cachedpath.WithMaxRetries(5),
	)
	if err != nil {
		t.Errorf("CachedPath with multiple options failed: %v", err)
	}

	if path != tmpFile.Name() {
		t.Errorf("CachedPath returned wrong path: %s", path)
	}
}
