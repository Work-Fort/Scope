// internal/infra/fortconfig/registry.go
package fortconfig

import (
	"sort"

	"github.com/spf13/viper"

	"github.com/Work-Fort/Scope/internal/domain"
)

var _ domain.FortRegistry = (*Registry)(nil)

type Registry struct{}

func New() *Registry {
	return &Registry{}
}

func (r *Registry) Forts() []domain.Fort {
	fortsMap := viper.GetStringMap("forts")
	forts := make([]domain.Fort, 0, len(fortsMap))
	for name := range fortsMap {
		if domain.ValidFortName(name) {
			forts = append(forts, r.readFort(name))
		}
	}
	sort.Slice(forts, func(i, j int) bool {
		return forts[i].Name < forts[j].Name
	})
	return forts
}

func (r *Registry) Fort(name string) (domain.Fort, bool) {
	if !domain.ValidFortName(name) {
		return domain.Fort{}, false
	}
	fortsMap := viper.GetStringMap("forts")
	if _, exists := fortsMap[name]; !exists {
		return domain.Fort{}, false
	}
	return r.readFort(name), true
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
