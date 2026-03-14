// internal/infra/fortconfig/registry_test.go
package fortconfig_test

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/Work-Fort/Scope/internal/domain"
	"github.com/Work-Fort/Scope/internal/infra/fortconfig"
)

func TestRegistry_Fort(t *testing.T) {
	viper.Reset()
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services", []map[string]any{
		{"url": "http://127.0.0.1:3000"},
		{"url": "http://127.0.0.1:9600"},
	})

	reg := fortconfig.New()
	fort, ok := reg.Fort("local")
	if !ok {
		t.Fatal("expected fort 'local' to exist")
	}
	if fort.Name != "local" {
		t.Errorf("got name %q, want %q", fort.Name, "local")
	}
	if len(fort.Services) != 2 {
		t.Errorf("got %d services, want 2", len(fort.Services))
	}

	_, ok = reg.Fort("nonexistent")
	if ok {
		t.Error("expected nonexistent fort to return false")
	}
}

func TestValidFortName(t *testing.T) {
	valid := []string{"local", "acme-corp", "a", "1", "a1", "test-fort-123"}
	invalid := []string{"-leading", "trailing-", "UPPER", "has space", "has/slash", "has..dots", ""}

	for _, name := range valid {
		if !domain.ValidFortName(name) {
			t.Errorf("expected %q to be valid", name)
		}
	}
	for _, name := range invalid {
		if domain.ValidFortName(name) {
			t.Errorf("expected %q to be invalid", name)
		}
	}
}
