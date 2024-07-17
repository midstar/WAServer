package main

import (
	"bytes"
	"encoding/json"
	"flag"
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

	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
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
		_, err := client.Get(baseURL)
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

func getObject(t *testing.T, path string, recv_obj interface{}) {
	t.Helper()
	resp, err := http.Get(fmt.Sprintf("%s/%s", baseURL, path))
	if err != nil {
		t.Fatalf("Unable to get path %s. Reason: %s", path, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code for path %s: %d (%s)",
			path, resp.StatusCode, respToString(resp.Body))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unable to read body for %s. Reason: %s", path, err)
	}
	err = json.Unmarshal(body, &recv_obj)
	if err != nil {
		t.Fatalf("Unable decode path %s. Reason: %s", path, err)
	}
}

func postObject(t *testing.T, path string, recv_obj interface{}, send_obj interface{}) {
	t.Helper()
	send_bytes, err := json.Marshal(send_obj)
	if err != nil {
		t.Fatalf("Unable encode object. Reason: %s", err)
	}
	var resp *http.Response
	resp, err = http.Post(fmt.Sprintf("%s/%s", baseURL, path),
		"application/json", bytes.NewBuffer(send_bytes))
	if err != nil {
		t.Fatalf("Unable to get path %s. Reason: %s", path, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code for path %s: %d (%s)",
			path, resp.StatusCode, respToString(resp.Body))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Unable to read body for %s. Reason: %s", path, err)
	}
	err = json.Unmarshal(body, &recv_obj)
	if err != nil {
		t.Fatalf("Unable decode path %s. Reason: %s", path, err)
	}
}

func TestStaticGet(t *testing.T) {
	startServer(t)

	html := getHTML(t, "app/3_in_a_row/index.html")
	if !strings.Contains(html, "<title>3 in a row</title>") {
		t.Fatal("Index html title missing")
	}

	shutdownServer(t)
}

func TestDataGet(t *testing.T) {
	startServer(t)

	send_obj := map[string]int{"foo": 1, "bar": 2}
	var recv_obj map[string]string
	postObject(t, "data/hej", &recv_obj, &send_obj)
	_, hasMessageKey := recv_obj["message"]
	assertTrue(t, "Key: message not defined", hasMessageKey)
	assertEqualsStr(t, "invalid message", "Data post", recv_obj["message"])

	shutdownServer(t)
}

func TestDataPost(t *testing.T) {
	startServer(t)

	var data map[string]string
	getObject(t, "data/hej", &data)
	_, hasMessageKey := data["message"]
	assertTrue(t, "Key: message not defined", hasMessageKey)
	assertEqualsStr(t, "invalid message", "Data get", data["message"])

	shutdownServer(t)
}
