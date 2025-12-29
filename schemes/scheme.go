package schemes

import "io"

// SchemeClient is the interface that all scheme clients must implement
type SchemeClient interface {
	// GetResource downloads the resource and writes to the writer
	GetResource(url string, writer io.Writer, headers map[string]string) error

	// GetSize returns the resource size in bytes
	GetSize(url string, headers map[string]string) (int64, error)

	// GetETag returns the resource ETag (for versioning)
	GetETag(url string, headers map[string]string) (string, error)

	// Scheme returns the scheme name (http, https, s3, gs, etc)
	Scheme() string
}

// Registry maintains a registry of scheme clients
var registry = make(map[string]SchemeClient)

// Register registers a new scheme client
func Register(client SchemeClient) {
	registry[client.Scheme()] = client
}

// GetClient gets a scheme client by name
func GetClient(scheme string) (SchemeClient, bool) {
	client, ok := registry[scheme]
	return client, ok
}

// GetSupportedSchemes retorna lista de esquemas suportados
func GetSupportedSchemes() []string {
	schemes := make([]string, 0, len(registry))
	for scheme := range registry {
		schemes = append(schemes, scheme)
	}
	return schemes
}

func init() {
	// Register default clients
	Register(NewHTTPClient())
}
