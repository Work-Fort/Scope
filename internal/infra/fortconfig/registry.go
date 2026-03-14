// internal/infra/fortconfig/registry.go
package fortconfig

import (
	"fmt"
	"sort"

	"github.com/spf13/viper"

	"github.com/Work-Fort/Scope/internal/domain"
)

// Compile-time interface check.
var _ domain.FortRegistry = (*Registry)(nil)

// Registry reads fort configuration from Viper.
type Registry struct{}

// New creates a new fort config registry.
func New() *Registry {
	return &Registry{}
}

// Forts returns all configured forts, sorted by name.
func (r *Registry) Forts() []domain.Fort {
	fortsMap := viper.GetStringMap("forts")
	forts := make([]domain.Fort, 0, len(fortsMap))
	for name := range fortsMap {
		forts = append(forts, r.readFort(name))
	}
	sort.Slice(forts, func(i, j int) bool {
		return forts[i].Name < forts[j].Name
	})
	return forts
}

// Active returns the currently active fort.
func (r *Registry) Active() domain.Fort {
	name := viper.GetString("active-fort")
	return r.readFort(name)
}

// SetActive switches the active fort. Returns an error if the fort does not exist.
func (r *Registry) SetActive(name string) error {
	fortsMap := viper.GetStringMap("forts")
	if _, ok := fortsMap[name]; !ok {
		return fmt.Errorf("fortconfig: fort %q not found", name)
	}
	viper.Set("active-fort", name)
	return nil
}

func (r *Registry) readFort(name string) domain.Fort {
	prefix := "forts." + name

	fort := domain.Fort{
		Name:    name,
		Local:   viper.GetBool(prefix + ".local"),
		Gateway: viper.GetString(prefix + ".gateway"),
	}

	var svcs []struct {
		URL string `mapstructure:"url"`
	}
	if err := viper.UnmarshalKey(prefix+".services", &svcs); err == nil {
		for _, s := range svcs {
			fort.Services = append(fort.Services, domain.ConfigService{URL: s.URL})
		}
	}

	return fort
}
