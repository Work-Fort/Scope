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
	cmd.Flags().IntVar(&port, "port", 8080, "Listen port")
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

	// Find the auth service URL for BFF token conversion.
	var authURL string
	for _, svc := range fort.Services {
		if svc.Name == "auth" && svc.Enabled {
			authURL = svc.URL
			break
		}
	}

	var tc *httpapi.TokenConverter
	if authURL != "" {
		tc = httpapi.NewTokenConverter(authURL)
	} else {
		log.Warn("auth service not configured — BFF token conversion disabled")
	}

	// SPA handler.
	var spaFS fs.FS
	if !dev {
		sub, err := fs.Sub(webFS, "placeholder")
		if err != nil {
			return fmt.Errorf("embedded SPA: %w", err)
		}
		spaFS = sub
	}

	handler := httpapi.NewHandler(fort, tc, spaFS)

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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
