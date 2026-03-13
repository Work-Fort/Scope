// internal/infra/fortconfig/registry_test.go
package fortconfig_test

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/Work-Fort/Scope/internal/domain"
	"github.com/Work-Fort/Scope/internal/infra/fortconfig"
)

func setupViper(t *testing.T) {
	t.Helper()
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services.sharkfin.url", "http://127.0.0.1:16000")
	viper.Set("forts.local.services.sharkfin.enabled", true)
	viper.Set("forts.local.services.sharkfin.ws-paths", []string{"/ws", "/presence"})
	viper.Set("forts.local.services.nexus.url", "http://127.0.0.1:9600")
	viper.Set("forts.local.services.nexus.enabled", true)
	viper.Set("forts.local.services.auth.url", "http://127.0.0.1:3000")
	viper.Set("forts.local.services.auth.enabled", true)
}

func TestActive(t *testing.T) {
	setupViper(t)

	reg := fortconfig.New()
	fort := reg.Active()

	if fort.Name != "local" {
		t.Fatalf("expected fort name 'local', got %q", fort.Name)
	}
	if !fort.Local {
		t.Fatal("expected fort to be local")
	}
	if len(fort.Services) != 3 {
		t.Fatalf("expected 3 services, got %d", len(fort.Services))
	}

	// Find sharkfin
	var sf *domain.Service
	for i := range fort.Services {
		if fort.Services[i].Name == "sharkfin" {
			sf = &fort.Services[i]
			break
		}
	}
	if sf == nil {
		t.Fatal("sharkfin service not found")
	}
	if sf.URL != "http://127.0.0.1:16000" {
		t.Fatalf("expected sharkfin URL http://127.0.0.1:16000, got %q", sf.URL)
	}
	if !sf.Enabled {
		t.Fatal("expected sharkfin to be enabled")
	}
	if len(sf.WSPaths) != 2 || sf.WSPaths[0] != "/ws" || sf.WSPaths[1] != "/presence" {
		t.Fatalf("unexpected ws-paths: %v", sf.WSPaths)
	}
}

func TestForts(t *testing.T) {
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services.auth.url", "http://127.0.0.1:3000")
	viper.Set("forts.local.services.auth.enabled", true)

	viper.Set("forts.remote.local", false)
	viper.Set("forts.remote.gateway", "https://fort.acme.com")
	viper.Set("forts.remote.services.auth.enabled", true)

	reg := fortconfig.New()
	forts := reg.Forts()

	if len(forts) != 2 {
		t.Fatalf("expected 2 forts, got %d", len(forts))
	}

	// Find the remote fort
	var remote *domain.Fort
	for i := range forts {
		if forts[i].Name == "remote" {
			remote = &forts[i]
			break
		}
	}
	if remote == nil {
		t.Fatal("remote fort not found")
	}
	if remote.Local {
		t.Fatal("expected remote fort to not be local")
	}
	if remote.Gateway != "https://fort.acme.com" {
		t.Fatalf("expected gateway https://fort.acme.com, got %q", remote.Gateway)
	}
}

func TestSetActive_Valid(t *testing.T) {
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services.auth.url", "http://127.0.0.1:3000")
	viper.Set("forts.local.services.auth.enabled", true)

	viper.Set("forts.remote.local", false)
	viper.Set("forts.remote.gateway", "https://fort.acme.com")
	viper.Set("forts.remote.services.auth.enabled", true)

	reg := fortconfig.New()
	if err := reg.SetActive("remote"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fort := reg.Active()
	if fort.Name != "remote" {
		t.Fatalf("expected active fort 'remote', got %q", fort.Name)
	}
}

func TestSetActive_Invalid(t *testing.T) {
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services.auth.url", "http://127.0.0.1:3000")
	viper.Set("forts.local.services.auth.enabled", true)

	reg := fortconfig.New()
	err := reg.SetActive("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent fort")
	}
}
