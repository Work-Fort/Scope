// internal/infra/fortconfig/registry_test.go
package fortconfig_test

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/Work-Fort/Scope/internal/infra/fortconfig"
)

func setupViper(t *testing.T) {
	t.Helper()
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services", []map[string]string{
		{"url": "http://127.0.0.1:16000"},
		{"url": "http://127.0.0.1:9600"},
		{"url": "http://127.0.0.1:3000"},
	})
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

	urls := make(map[string]bool)
	for _, svc := range fort.Services {
		urls[svc.URL] = true
	}
	if !urls["http://127.0.0.1:16000"] {
		t.Fatal("missing URL 16000")
	}
	if !urls["http://127.0.0.1:9600"] {
		t.Fatal("missing URL 9600")
	}
	if !urls["http://127.0.0.1:3000"] {
		t.Fatal("missing URL 3000")
	}
}

func TestForts(t *testing.T) {
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services", []map[string]string{
		{"url": "http://127.0.0.1:3000"},
	})

	viper.Set("forts.remote.local", false)
	viper.Set("forts.remote.gateway", "https://fort.acme.com")
	viper.Set("forts.remote.services", []map[string]string{
		{"url": "https://fort.acme.com/api/auth"},
	})

	reg := fortconfig.New()
	forts := reg.Forts()

	if len(forts) != 2 {
		t.Fatalf("expected 2 forts, got %d", len(forts))
	}

	// Forts are sorted by name, so local comes first, remote second.
	if forts[0].Name != "local" {
		t.Fatalf("expected first fort 'local', got %q", forts[0].Name)
	}
	if forts[1].Name != "remote" {
		t.Fatalf("expected second fort 'remote', got %q", forts[1].Name)
	}
	if forts[1].Local {
		t.Fatal("expected remote fort to not be local")
	}
	if forts[1].Gateway != "https://fort.acme.com" {
		t.Fatalf("expected gateway https://fort.acme.com, got %q", forts[1].Gateway)
	}
}

func TestSetActive_Valid(t *testing.T) {
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services", []map[string]string{
		{"url": "http://127.0.0.1:3000"},
	})

	viper.Set("forts.remote.local", false)
	viper.Set("forts.remote.gateway", "https://fort.acme.com")
	viper.Set("forts.remote.services", []map[string]string{
		{"url": "https://fort.acme.com/api/auth"},
	})

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
	viper.Set("forts.local.services", []map[string]string{
		{"url": "http://127.0.0.1:3000"},
	})

	reg := fortconfig.New()
	err := reg.SetActive("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent fort")
	}
}
