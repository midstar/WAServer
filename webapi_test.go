package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

var baseURL = "http://localhost:8080"

func startServer(t *testing.T) {
	t.Helper()
	go main()
	waitServer(t)
}

// waitServer waits for the server to be up and running
func waitServer(t *testing.T) {
	t.Helper()
	client := http.Client{Timeout: 100 * time.Millisecond}
	maxTries := 50
	for i := 0; i < maxTries; i++ {
		_, err := client.Get(fmt.Sprintf("%s", baseURL))
		if err == nil {
			// Up and running :-)
			return
		}
	}
	t.Fatalf("Server never started")
}

// shutdownServer shuts down server and clears the serveMux
func shutdownServer(t *testing.T) {
	// No answer expected on POST shutdown (short timeout)
	client := http.Client{Timeout: 1 * time.Second}
	client.Post(fmt.Sprintf("%s/service/shutdown", baseURL), "", nil)

	// Reset the serveMux
	http.DefaultServeMux = new(http.ServeMux)
}

func TestExample(t *testing.T) {
	startServer(t)
	shutdownServer(t)
}
