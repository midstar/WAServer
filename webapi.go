package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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

func (wa *WebAPI) handleDataGet(w http.ResponseWriter, r *http.Request) {
	dir, file, _ := dirAndJsonFile(r.URL.Path)
	// Tests shows that Golang server don't allow invalid paths, thus
	// no error needs to be handled
	if file == "" {
		messageResponse(w, http.StatusNotImplemented, "Directory access not implemented")
		return
	}
	fullPath := path.Join(wa.dataPath, dir, file)
	dat, err := os.ReadFile(fullPath)
	if err != nil {
		messageResponse(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (wa *WebAPI) handleDataPost(w http.ResponseWriter, r *http.Request) {
	messageResponse(w, http.StatusOK, "Data post")
}

func (wa *WebAPI) handleShutdown(w http.ResponseWriter, r *http.Request) {
	wa.Stop()
}

func messageResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	jsonResponse := fmt.Sprintf(`{"message": "%s"}`, message)
	w.Write([]byte(jsonResponse))
}

// Interpretets URL.path and assumes paths ending with / is
// a directory and last entity of path (which is not ending
// with /) is a json file.
func dirAndJsonFile(urlPath string) (string, string, error) {
	const prefix = "/data/"
	if !strings.HasPrefix(urlPath, prefix) {
		return "", "", fmt.Errorf("path needs to start with %s but was: %s", prefix, urlPath)
	}
	urlPath = strings.TrimPrefix(urlPath, prefix)
	dir, file := path.Split(urlPath)

	dir = path.Clean(dir)
	if strings.HasPrefix(dir, "..") {
		return "", "", fmt.Errorf("hacker attack. Someone tries to access: %s", dir)
	}

	if file != "" {
		file = file + ".json"
	}

	return dir, file, nil
}
