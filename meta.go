package cachedpath

import (
	"encoding/json"
	"os"
	"time"
)

// Meta armazena metadados sobre arquivos em cache
type Meta struct {
	URL        string    `json:"url"`
	ETag       string    `json:"etag"`
	CachedPath string    `json:"cached_path"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewMeta creates a new Meta instance
func NewMeta(url, cachedPath, etag string) *Meta {
	return &Meta{
		URL:        url,
		ETag:       etag,
		CachedPath: cachedPath,
		CreatedAt:  time.Now(),
	}
}

// SaveToFile saves metadata to a file
func (m *Meta) SaveToFile(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadMetaFromFile loads metadata from a file
func LoadMetaFromFile(path string) (*Meta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}
