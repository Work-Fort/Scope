package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Work-Fort/Scope/internal/infra/httpapi"
)

func TestTracker_InitialProbe(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":   "ok",
			"name":     "sharkfin",
			"label":    "Chat",
			"route":    "/chat",
			"ws_paths": []string{"/ws"},
		})
	}))
	defer svc.Close()

	tracker := httpapi.NewServiceTracker([]string{svc.URL})
	tracker.InitialProbe(context.Background())

	services := tracker.Services()
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
	if services[0].Name != "sharkfin" {
		t.Fatalf("expected name sharkfin, got %q", services[0].Name)
	}
	if services[0].Label != "Chat" {
		t.Fatalf("expected label Chat, got %q", services[0].Label)
	}
	if !services[0].UI {
		t.Fatal("expected ui=true")
	}
	if services[0].Connected {
		t.Fatal("expected connected=false for WS service before any connections")
	}
}

func TestTracker_ConflictDetection(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"name":   "sharkfin",
			"label":  "Chat",
			"route":  "/chat",
		})
	})

	svc1 := httptest.NewServer(handler)
	defer svc1.Close()
	svc2 := httptest.NewServer(handler)
	defer svc2.Close()

	tracker := httpapi.NewServiceTracker([]string{svc1.URL, svc2.URL})
	tracker.InitialProbe(context.Background())

	services := tracker.Services()
	if len(services) != 1 {
		t.Fatalf("expected 1 service (first wins), got %d", len(services))
	}
	if services[0].URL != svc1.URL {
		t.Fatalf("expected first URL to win, got %q", services[0].URL)
	}

	conflicts := tracker.Conflicts()
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].URL != svc2.URL {
		t.Fatalf("expected conflict URL to be svc2, got %q", conflicts[0].URL)
	}
}

func TestTracker_UnreachableService(t *testing.T) {
	tracker := httpapi.NewServiceTracker([]string{"http://127.0.0.1:1"})
	tracker.InitialProbe(context.Background())

	if len(tracker.Services()) != 0 {
		t.Fatalf("expected 0 services for unreachable URL, got %d", len(tracker.Services()))
	}
}

func TestTracker_NoUIService(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "unavailable",
			"name":   "auth",
			"label":  "Auth",
			"route":  "/auth",
		})
	}))
	defer svc.Close()

	tracker := httpapi.NewServiceTracker([]string{svc.URL})
	tracker.InitialProbe(context.Background())

	services := tracker.Services()
	if len(services) != 1 {
		t.Fatalf("expected 1 service (503 still registers), got %d", len(services))
	}
	if services[0].Name != "auth" {
		t.Fatalf("expected name auth, got %q", services[0].Name)
	}
	if services[0].UI {
		t.Fatal("expected ui=false for 503 service")
	}
	if !services[0].Connected {
		t.Fatal("expected connected=true for non-WS reachable service")
	}
}

func TestTracker_WSConnectionTracking(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":   "ok",
			"name":     "sharkfin",
			"label":    "Chat",
			"route":    "/chat",
			"ws_paths": []string{"/ws"},
		})
	}))
	defer svc.Close()

	tracker := httpapi.NewServiceTracker([]string{svc.URL})
	tracker.InitialProbe(context.Background())

	if tracker.Services()[0].Connected {
		t.Fatal("WS service should start disconnected")
	}

	tracker.OnConnect("sharkfin")
	if !tracker.Services()[0].Connected {
		t.Fatal("expected connected after OnConnect")
	}

	tracker.OnConnect("sharkfin")
	tracker.OnDisconnect("sharkfin")
	if !tracker.Services()[0].Connected {
		t.Fatal("expected still connected (ref count = 1)")
	}

	tracker.OnDisconnect("sharkfin")
	if tracker.Services()[0].Connected {
		t.Fatal("expected disconnected (ref count = 0)")
	}
}

func TestTracker_BackgroundPolling(t *testing.T) {
	var probeCount int32
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&probeCount, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"name":   "nexus",
			"label":  "Nexus",
			"route":  "/nexus",
		})
	}))
	defer svc.Close()

	tracker := httpapi.NewServiceTracker([]string{svc.URL})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracker.InitialProbe(ctx)
	tracker.StartPolling(ctx, 50*time.Millisecond)

	time.Sleep(200 * time.Millisecond)
	cancel()

	if c := atomic.LoadInt32(&probeCount); c < 3 {
		t.Fatalf("expected at least 3 probes, got %d", c)
	}
}

func TestTracker_ServiceComesBackUp(t *testing.T) {
	var respondOK int32

	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&respondOK) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"name":   "nexus",
			"label":  "Nexus",
			"route":  "/nexus",
		})
	}))
	defer svc.Close()

	var discovered int32
	tracker := httpapi.NewServiceTracker([]string{svc.URL})
	tracker.OnServiceDiscovered = func(svc httpapi.TrackedService) {
		atomic.AddInt32(&discovered, 1)
	}

	tracker.InitialProbe(context.Background())
	if len(tracker.Services()) != 0 {
		t.Fatal("expected 0 services while down")
	}

	atomic.StoreInt32(&respondOK, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tracker.StartPolling(ctx, 50*time.Millisecond)

	time.Sleep(200 * time.Millisecond)
	cancel()

	if len(tracker.Services()) != 1 {
		t.Fatalf("expected 1 service after coming up, got %d", len(tracker.Services()))
	}
	if atomic.LoadInt32(&discovered) != 1 {
		t.Fatal("expected OnServiceDiscovered to fire once")
	}
}

func TestTracker_ServiceByName(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"name":   "nexus",
			"label":  "Nexus",
			"route":  "/nexus",
		})
	}))
	defer svc.Close()

	tracker := httpapi.NewServiceTracker([]string{svc.URL})
	tracker.InitialProbe(context.Background())

	s, ok := tracker.ServiceByName("nexus")
	if !ok {
		t.Fatal("expected to find nexus")
	}
	if s.Name != "nexus" {
		t.Fatalf("expected name nexus, got %q", s.Name)
	}

	_, ok = tracker.ServiceByName("nonexistent")
	if ok {
		t.Fatal("expected not found for nonexistent")
	}
}
