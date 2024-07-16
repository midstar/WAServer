package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strings"
)

// WebAPI represents the REST API server.
type WebAPI struct {
	server      *http.Server
	appPath     string // Path to the applications
	dataPath    string // Path to the data
	tlsCertFile string // TLS certification file ("" means no TLS)
	tlsKeyFile  string // TLS key file ("" means no TLS)
}

// CreateWebAPI creates a new Web API instance
func CreateWebAPI(port int, appPath, dataPath string,
	tlsCertFile, tlsKeyFile string) *WebAPI {
	portStr := fmt.Sprintf(":%d", port)
	server := &http.Server{Addr: portStr}
	webAPI := &WebAPI{
		server:      server,
		appPath:     appPath,
		dataPath:    dataPath,
		tlsCertFile: tlsCertFile,
		tlsKeyFile:  tlsKeyFile}
	http.Handle("/app/", http.StripPrefix("/app/",
		http.FileServer(http.Dir(appPath))))
	http.HandleFunc("GET /data/", webAPI.handleDataGet)
	http.HandleFunc("POST /data/", webAPI.handleDataPost)
	http.HandleFunc("POST /service/shutdown", webAPI.handleShutdown)
	return webAPI
}

func (wa *WebAPI) Start() chan bool {
	done := make(chan bool)

	go func() {
		slog.Info(fmt.Sprintf("Serving path %s on port %s", wa.appPath, wa.server.Addr))
		if wa.tlsCertFile != "" && wa.tlsKeyFile != "" {
			slog.Info("Using TLS (HTTPS)")
			if err := wa.server.ListenAndServeTLS(wa.tlsCertFile, wa.tlsKeyFile); err != nil {
				// cannot panic, because this probably is an intentional close
				slog.Info(fmt.Sprintf("WebAPI: ListenAndServeTLS() shutdown reason: %s", err))
			}
		} else {
			if err := wa.server.ListenAndServe(); err != nil {
				// cannot panic, because this probably is an intentional close
				slog.Info(fmt.Sprintf("WebAPI: ListenAndServe() shutdown reason: %s", err))
			}
		}
		done <- true // Signal that http server has stopped
	}()
	return done
}

// Stop stops the HTTP server.
func (wa *WebAPI) Stop() {
	wa.server.Shutdown(context.Background())
}

func (wa *WebAPI) handleDataPost(w http.ResponseWriter, r *http.Request) {
	/*var head string
	originalURL := r.URL.Path
	llog.Trace("Got request: %s", r.URL.Path)
	head, r.URL.Path = shiftPath(r.URL.Path)*/

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Data post"}`))
}

func (wa *WebAPI) handleDataGet(w http.ResponseWriter, r *http.Request) {
	/*var head string
	originalURL := r.URL.Path
	llog.Trace("Got request: %s", r.URL.Path)
	head, r.URL.Path = shiftPath(r.URL.Path)*/

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Data get"}`))
}

func (wa *WebAPI) handleShutdown(w http.ResponseWriter, r *http.Request) {
	wa.Stop()
}

// shiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
