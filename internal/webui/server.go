// Package webui serves an interactive HTTP UI for execute_query.
package webui

import (
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"strings"

	executequery "github.com/apstndb/go-googlesql-executequery"
)

// Server holds the state for the web UI.
type Server struct {
	port int
}

// NewServer creates a web UI server listening on port.
func NewServer(port int) *Server {
	return &Server{port: port}
}

// Handler returns the HTTP handler used by Run (for tests and embedding).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/run", s.handleRun)
	return mux
}

// Run starts the HTTP server and blocks until it shuts down.
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Listening on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, s.Handler())
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := tmpl.Execute(w, pageData{
		Style:     template.CSS(pageStyleCSS),
		indexData: defaultIndexData(),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ct, _, perr := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if perr != nil && r.Header.Get("Content-Type") != "" {
		http.Error(w, perr.Error(), http.StatusBadRequest)
		return
	}
	switch ct {
	case "multipart/form-data":
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	sql := strings.TrimSpace(r.FormValue("query"))
	if sql == "" {
		http.Error(w, "no SQL provided", http.StatusBadRequest)
		return
	}

	cfg, err := configFromForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var hw htmlWriter
	if err := executequery.Run(r.Context(), sql, cfg, &hw); err != nil {
		hw.setError(err.Error())
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(hw.Bytes()); err != nil {
		_ = err
	}
}
