package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	/*var head string
	originalURL := r.URL.Path
	llog.Trace("Got request: %s", r.URL.Path)
	head, r.URL.Path = shiftPath(r.URL.Path)*/

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "hello world"}`))
}

func startServer(port int, appDir, dataDir string, tlsEnable bool, tlsCertFile, tlsKeyFile string) {
	http.Handle("/app/", http.StripPrefix("/app/",
		http.FileServer(http.Dir(appDir))))
	http.Handle("/", http.RedirectHandler("/app/", 302))
	http.HandleFunc("/api/", handleRequest)

	slog.Debug("This is a debug message")

	if tlsEnable {
		slog.Info(fmt.Sprintf("Serving path %s on port %d over HTTPS\n", appDir, port))

		err := http.ListenAndServeTLS(":"+strconv.Itoa(port), tlsCertFile, tlsKeyFile, nil)
		if err != nil {
			slog.Error(fmt.Sprintf("Error! %s\n", err))
		}
	} else {
		slog.Info(fmt.Sprintf("Serving path %s on port %d over HTTP\n", appDir, port))

		err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
		if err != nil {
			slog.Error(fmt.Sprintf("Error! %s\n", err))
		}
	}
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
