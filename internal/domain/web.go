// internal/domain/web.go
package domain

// Fort is a named collection of services. Users can belong to multiple
// forts and switch between them.
type Fort struct {
	// Name is the fort identifier (e.g., "local", "acme-corp").
	Name string

	// Local controls how the proxy routes traffic.
	// true = proxy directly to each service's URL.
	// false = proxy through Gateway.
	Local bool

	// Gateway is the single origin URL for remote forts.
	// Only used when Local is false.
	Gateway string

	// Services lists the backend services in this fort.
	Services []Service
}

// Service is a backend service in a fort.
type Service struct {
	// Name is the service identifier (e.g., "auth", "sharkfin", "nexus", "hive").
	Name string

	// URL is the direct backend URL (e.g., "http://127.0.0.1:16000").
	// Only used when the fort's Local flag is true.
	URL string

	// WSPaths is a whitelist of paths that accept WebSocket upgrade.
	// Matched against the path suffix after the /api/{service} prefix is stripped.
	// Example: ["/ws", "/presence"]
	WSPaths []string

	// Enabled controls whether the proxy accepts requests for this service.
	// Disabled services return 503.
	Enabled bool
}

// FortRegistry reads fort configuration.
type FortRegistry interface {
	// Forts returns all configured forts.
	Forts() []Fort

	// Active returns the currently active fort.
	Active() Fort

	// SetActive switches the active fort.
	// Returns an error if the fort name does not exist.
	SetActive(name string) error
}
