// internal/domain/web.go
package domain

import "regexp"

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
	Services []ConfigService
}

// ConfigService is what comes from the fort config file — just a URL.
// All configured services are considered enabled. To disable a service,
// remove it from the config.
type ConfigService struct {
	URL string
}

// FortRegistry provides access to configured forts.
type FortRegistry interface {
	// Forts returns all configured forts.
	Forts() []Fort

	// Fort looks up a fort by name.
	// Returns (fort, true) if found, (zero value, false) if not.
	Fort(name string) (Fort, bool)
}

var fortNameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// ValidFortName returns true if name is a valid fort identifier.
// Lowercase alphanumeric + hyphens, no leading/trailing hyphens.
func ValidFortName(name string) bool {
	return fortNameRe.MatchString(name)
}
