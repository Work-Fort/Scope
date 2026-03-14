package web

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"charm.land/log/v2"
	"github.com/spf13/cobra"

	"github.com/Work-Fort/Scope/internal/infra/fortconfig"
	"github.com/Work-Fort/Scope/internal/infra/httpapi"
)

var (
	bind   string
	port   int
	dev    bool
	devURL string
	noOpen bool
)

// New creates the "web" subcommand.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Start the web UI server",
		Long:  "Serve the WorkFort web shell and proxy requests to backend services.",
		RunE:  run,
	}

	cmd.Flags().StringVar(&bind, "bind", "127.0.0.1", "Listen address")
	cmd.Flags().IntVar(&port, "port", 16100, "Listen port")
	cmd.Flags().BoolVar(&dev, "dev", false, "Proxy SPA to Vite dev server")
	cmd.Flags().StringVar(&devURL, "dev-url", "http://localhost:5173", "Vite dev server URL (used with --dev)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Don't auto-open browser")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	registry := fortconfig.New()
	fort := registry.Active()

	log.Info("starting web server",
		"fort", fort.Name,
		"local", fort.Local,
		"services", len(fort.Services),
	)

	// Create signal context early — used by tracker and shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create service tracker and run initial probe.
	urls := make([]string, len(fort.Services))
	for i, svc := range fort.Services {
		urls[i] = svc.URL
	}

	tracker := httpapi.NewServiceTracker(urls)
	tracker.InitialProbe(ctx)
	tracker.StartPolling(ctx, 10*time.Second)

	// Find the auth service URL from tracker (discovered via /ui/health).
	var tc *httpapi.TokenConverter
	if authSvc, ok := tracker.ServiceByName("auth"); ok {
		tc = httpapi.NewTokenConverter(authSvc.URL)
	} else {
		log.Warn("auth service not discovered — BFF token conversion disabled")
	}

	// SPA handler.
	var spaFS fs.FS
	if !dev {
		sub, err := fs.Sub(webFS, "dist")
		if err != nil {
			return fmt.Errorf("embedded SPA: %w", err)
		}
		spaFS = sub
	}

	handler := httpapi.NewHandler(fort, tracker, tc, spaFS)

	// In dev mode, wrap the handler to proxy non-/api/* to Vite.
	if dev {
		devProxy := httpapi.NewSPADevProxy(devURL)
		topMux := http.NewServeMux()
		topMux.Handle("/api/", handler)
		topMux.Handle("/", devProxy)
		handler = topMux
	}

	addr := fmt.Sprintf("%s:%d", bind, port)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Graceful shutdown.
	go func() {
		<-ctx.Done()
		log.Info("shutting down web server")
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutCtx)
	}()

	url := fmt.Sprintf("http://%s", addr)
	log.Info("web server listening", "url", url)

	if !noOpen {
		go openBrowser(url)
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("web server: %w", err)
	}

	return nil
}

func openBrowser(url string) {
	// Small delay to let the server start.
	time.Sleep(200 * time.Millisecond)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}
