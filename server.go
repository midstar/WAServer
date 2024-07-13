package main

import (
    "net/http"
	"strconv"
	"fmt"
	"os"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"message": "hello world"}`))
}

func startServer(port int, appDir, dataDir string, tlsEnable bool, tlsCertFile, tlsKeyFile string) {
	http.Handle("/app/", http.StripPrefix("/app/", 
		http.FileServer(http.Dir(appDir))))
	http.Handle("/", http.RedirectHandler("/app/", 302))
	http.HandleFunc("/api/", handleRequest)
	
	if tlsEnable {
		fmt.Printf("Serving path %s on port %d over HTTPS\n", appDir, port)

		err := http.ListenAndServeTLS(":"+strconv.Itoa(port), tlsCertFile, tlsKeyFile, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error! %s\n", err)
		}
	} else {
		fmt.Printf("Serving path %s on port %d over HTTP\n", appDir, port)

		err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error! %s\n", err)
		}	
	} 
}