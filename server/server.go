package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/rustyeddy/otto/messanger"
	"github.com/rustyeddy/otto/utils"
)

// Server serves up HTTP on Addr (default 0.0.0.0:8011)
// It takes care of REST API, serving the web app if Appdir
// does not equal nil and initial Websocket upgrade
type Server struct {
	*http.Server       `json:"-"`
	*http.ServeMux     `json:"-"`
	*template.Template `json:"-"`

	EndPoints sync.Map `json:"routes"`
}

var (
	wserv  WServe
	server *Server
)

func GetServer() *Server {
	if server == nil {
		server = NewServer()
	}
	return server
}

func NewServer() *Server {
	s := &Server{
		Server: &http.Server{
			Addr: ":8011",
		},
	}
	s.ServeMux = http.NewServeMux()
	return s
}

// Register to handle HTTP requests for particular paths in the
// URL or MQTT channel.
func (s *Server) Register(p string, h http.Handler) error {

	if p == "" || h == nil {
		return errors.New("Server.Register can not have null path or handler")
	}

	// Check if already registered to avoid duplicate registration errors
	_, alreadyRegistered := s.EndPoints.Load(p)
	if alreadyRegistered {
		return nil // Already registered, skip
	}

	// get this to log to a file (or syslog) by default
	//	slog.Info("HTTP REST API Registered: ", "path", p)
	s.EndPoints.Store(p, h)
	s.Handle(p, h)
	return nil
}

// Start the HTTP server after registering REST API callbacks
// and initializing the Web application directory
func (s *Server) Start(done chan any) {
	s.Register("/ping", Ping{})
	s.Register("/api", s)
	s.Register("/api/topics", messanger.GetTopics())
	s.Register("/api/stats", utils.Stats{})

	slog.Info("Starting hub Web and REST server on ", "addr", s.Addr)
	go http.ListenAndServe(s.Addr, s.ServeMux)
	<-done
	s.Shutdown(context.Background())
}

func (s *Server) Appdir(path, file string) {
	slog.Info("appdir", "path", path)
	s.Register(path, http.FileServer(http.Dir(file)))
}

func (s *Server) EndPointCount() int {
	count := 0
	s.EndPoints.Range(func(k, v any) bool {
		count++
		return true
	})
	return count
}

func (s *Server) EmbedTempl(path string, fsys embed.FS, data any) {

	s.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path
		ext := filepath.Ext(url)

		switch ext {
		case ".css":
			w.Header().Set("Content-Type", "text/css")
			http.ServeFileFS(w, r, fsys, "app"+url)
			return

		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
			http.ServeFileFS(w, r, fsys, "app"+url)
			return

		default:
			var err error
			if s.Template == nil {
				s.Template, err = template.ParseFS(fsys, "app/*.html")
				if err != nil {
					slog.Error("Failed to parse web template: ", "error", err.Error())
					return
				}
			}
			s.Template.Execute(w, data)
		}
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ep := struct {
		Routes []string
	}{}
	s.EndPoints.Range(func(k, v any) bool {
		ep.Routes = append(ep.Routes, k.(string))
		return true
	})

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(ep)
	if err != nil {
		slog.Error("Server.ServeHTTP failed to encode", "error", err)
	}
}
