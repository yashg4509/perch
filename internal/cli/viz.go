package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/stackstatus"
	webdist "github.com/yashg4509/perch/web"
)

func newVizCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "viz",
		Short: "Start local web UI for stack visualization",
		RunE:  runViz,
	}
	cmd.Flags().Int("port", 3131, "HTTP listen port")
	return cmd
}

func runViz(cmd *cobra.Command, args []string) error {
	_ = args
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return err
	}
	defaultEnv, err := cmd.Flags().GetString("env")
	if err != nil {
		return err
	}
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port %d", port)
	}

	_, _, err = loadStackFromWD()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "# → Building graph from perch.yaml...")
	addr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "# → Perch UI running at %s\n", baseURL)
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "# → Opening browser...")

	go openBrowser(baseURL + "/")

	// go:embed all:dist exposes paths as dist/index.html, dist/assets/..., not index.html at root.
	uiFS, err := fs.Sub(webdist.Dist, "dist")
	if err != nil {
		return fmt.Errorf("web UI embed: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/graph", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		serveGraphJSON(w, r, defaultEnv)
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		serveStatusJSON(w, r, defaultEnv)
	})
	mux.Handle("/", spaHandler(uiFS))

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return srv.ListenAndServe()
}

func loadStackFromWD() (*config.Config, *provider.Registry, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}
	perchPath, err := config.FindPerchYAML(wd)
	if err != nil {
		return nil, nil, err
	}
	raw, err := os.ReadFile(perchPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read config: %w", err)
	}
	cfg, err := config.Load(raw)
	if err != nil {
		return nil, nil, err
	}
	root := filepath.Dir(perchPath)
	reg, err := loadRegistryForProject(root)
	if err != nil {
		return nil, nil, err
	}
	return cfg, reg, nil
}

func envFromRequest(r *http.Request, defaultEnv string) string {
	q := r.URL.Query().Get("env")
	if strings.TrimSpace(q) == "" {
		return defaultEnv
	}
	return q
}

func isBadEnvErr(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "unknown environment") ||
		strings.Contains(s, "environment name is required")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func serveGraphJSON(w http.ResponseWriter, r *http.Request, defaultEnv string) {
	env := envFromRequest(r, defaultEnv)
	cfg, reg, err := loadStackFromWD()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	g, err := graph.Build(cfg, reg, env)
	if err != nil {
		if isBadEnvErr(err) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rep := graph.NewJSONReport(g)
	writeJSON(w, http.StatusOK, rep)
}

func serveStatusJSON(w http.ResponseWriter, r *http.Request, defaultEnv string) {
	env := envFromRequest(r, defaultEnv)
	cfg, reg, err := loadStackFromWD()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ctx := context.Background()
	rep, err := stackstatus.Collect(ctx, cfg, env, reg)
	if err != nil {
		if isBadEnvErr(err) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

func spaHandler(root fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		upath := path.Clean(r.URL.Path)
		if upath == "." || upath == "/" {
			upath = "index.html"
		} else {
			upath = strings.TrimPrefix(upath, "/")
		}
		if _, err := fs.Stat(root, upath); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				upath = "index.html"
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		http.ServeFileFS(w, r, root, upath)
	})
}

func openBrowser(rawURL string) {
	var u *url.URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	s := u.String()
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", s).Start()
	case "windows":
		_ = exec.Command("cmd", "/c", "start", "", s).Start()
	default:
		_ = exec.Command("xdg-open", s).Start()
	}
}
