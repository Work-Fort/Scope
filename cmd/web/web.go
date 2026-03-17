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
	openBrowserFlag bool
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
	cmd.Flags().BoolVar(&openBrowserFlag, "open", false, "Auto-open browser on start")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	registry := fortconfig.New()
	forts := registry.Forts()
	if len(forts) == 0 {
		return fmt.Errorf("no forts configured")
	}

	var spaHandler http.Handler
	if dev {
		spaHandler = httpapi.NewSPADevProxy(devURL)
	} else {
		distFS, err := fs.Sub(webFS, "dist")
		if err != nil {
			return fmt.Errorf("embedded SPA: %w", err)
		}
		spaHandler = httpapi.NewSPAHandler(distFS)
	}

	router := httpapi.NewFortRouter(registry, spaHandler)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router.StartIdleCleanup(ctx, 30*time.Minute)

	addr := fmt.Sprintf("%s:%d", bind, port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   60 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MiB
	}

	url := fmt.Sprintf("http://%s", addr)
	log.Info("web server listening", "url", url, "forts", len(forts))

	if openBrowserFlag {
		go openBrowser(url)
	}

	go func() {
		<-ctx.Done()
		log.Info("shutting down web server")
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
