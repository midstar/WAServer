package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var baseURL = "http://localhost:8080"

func startServer(t *testing.T) {
	t.Helper()

	os.Args = []string{"test", "app"}
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

func respToString(response io.ReadCloser) string {
	defer response.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(response)
	return buf.String()
}

func getHTML(t *testing.T, path string) string {
	t.Helper()
	resp, err := http.Get(fmt.Sprintf("%s/%s", baseURL, path))
	assertExpectNoErr(t, "", err)
	assertEqualsInt(t, "", int(http.StatusOK), int(resp.StatusCode))
	assertEqualsStr(t, "", "text/html; charset=utf-8", resp.Header.Get("content-type"))
	defer resp.Body.Close()
	return respToString(resp.Body)
}

func TestGetStatic(t *testing.T) {
	startServer(t)

	html := getHTML(t, "app/3_in_a_row/index.html")
	if !strings.Contains(html, "<title>3 in a row</title>") {
		t.Fatal("Index html title missing")
	}

	shutdownServer(t)
}
