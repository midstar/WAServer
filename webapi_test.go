package main

import (
	"bytes"
	"crypto/tls"
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
	os.Args = []string{"test", "app", ".test/data"}
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

func getObject(t *testing.T, path string, expectedStatus int, recv_obj interface{}) {
	t.Helper()
	resp, err := http.Get(fmt.Sprintf("%s/%s", baseURL, path))
	if err != nil {
		t.Fatalf("Unable to get path %s. Reason: %s", path, err)
	}
	if resp.StatusCode != expectedStatus {
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
	defer shutdownServer(t)

	html := getHTML(t, "app/3_in_a_row/index.html")
	if !strings.Contains(html, "<title>3 in a row</title>") {
		t.Fatal("Index html title missing")
	}
}

func TestDataPost(t *testing.T) {
	startServer(t)
	defer shutdownServer(t)

	send_obj := map[string]int{"foo": 1, "bar": 2}
	var recv_obj map[string]string
	postObject(t, "data/hej", &recv_obj, &send_obj)
	_, hasMessageKey := recv_obj["message"]
	assertTrue(t, "Key: message not defined", hasMessageKey)
	assertEqualsStr(t, "invalid message", "Data post", recv_obj["message"])
}

func TestDataGet(t *testing.T) {
	startServer(t)
	defer shutdownServer(t)

	var data map[string]string

	// Test invalid URL
	getObject(t, "data/fileDontExist", http.StatusNotFound, &data)

	// Test file that exists
	getObject(t, "data/adir/myfile", http.StatusOK, &data)
	_, hasAkey := data["akey"]
	assertTrue(t, "Key: akey not defined", hasAkey)
	assertEqualsStr(t, "invalid message", "avalue", data["akey"])

}

func TestTLS(t *testing.T) {
	var baseHttpsURL = "https://localhost:8080"

	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"test", "-d", "-s", "-c", ".test/cert.pem",
		"-k", ".test/key.pem", "app", ".test/data"}
	go main()

	// Create the client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpsClient := &http.Client{Transport: tr, Timeout: 100 * time.Millisecond}

	// Wait until server goes up
	maxTries := 50
	i := 0
	for i = 0; i < maxTries; i++ {
		_, err := httpsClient.Get(baseHttpsURL)
		if err == nil {
			// Up and running :-)
			break
		}
	}
	assertTrue(t, "Server never started using TLS", i < maxTries)

	// Access the main app page
	resp, err := httpsClient.Get(fmt.Sprintf("%s/app", baseHttpsURL))
	assertExpectNoErr(t, "Unable to connect over TLS", err)
	defer resp.Body.Close()
	assertEqualsInt(t, "", int(http.StatusOK), int(resp.StatusCode))
	assertEqualsStr(t, "", "text/html; charset=utf-8", resp.Header.Get("content-type"))

	// Shutdown the server
	// No answer expected on POST shutdown (short timeout)
	httpsClient = &http.Client{Timeout: 1 * time.Second, Transport: tr}
	httpsClient.Post(fmt.Sprintf("%s/service/shutdown", baseHttpsURL), "", nil)

	// Reset the serveMux
	http.DefaultServeMux = new(http.ServeMux)

}

func TestDirAndJsonFile(t *testing.T) {
	dir, file, err := dirAndJsonFile("/nodata/afile")
	assertEqualsStr(t, "", "", dir)
	assertEqualsStr(t, "", "", file)
	assertExpectErr(t, "", err)

	dir, file, err = dirAndJsonFile("/data/")
	assertEqualsStr(t, "", ".", dir)
	assertEqualsStr(t, "", "", file)
	assertExpectNoErr(t, "", err)

	dir, file, err = dirAndJsonFile("/data/afile")
	assertEqualsStr(t, "", ".", dir)
	assertEqualsStr(t, "", "afile.json", file)
	assertExpectNoErr(t, "", err)

	dir, file, err = dirAndJsonFile("/data/adir/")
	assertEqualsStr(t, "", "adir", dir)
	assertEqualsStr(t, "", "", file)
	assertExpectNoErr(t, "", err)

	dir, file, err = dirAndJsonFile("/data/adir/asubdir/")
	assertEqualsStr(t, "", "adir/asubdir", dir)
	assertEqualsStr(t, "", "", file)
	assertExpectNoErr(t, "", err)

	dir, file, err = dirAndJsonFile("/data/adir/asubdir/file")
	assertEqualsStr(t, "", "adir/asubdir", dir)
	assertEqualsStr(t, "", "file.json", file)
	assertExpectNoErr(t, "", err)

	// Hacker attack to get access to files outside the
	// data directory
	dir, file, err = dirAndJsonFile("/data/../asubdir/file")
	assertEqualsStr(t, "", "", dir)
	assertEqualsStr(t, "", "", file)
	assertExpectErr(t, "", err)

	dir, file, err = dirAndJsonFile("/data/asubdir/../../file")
	assertEqualsStr(t, "", "", dir)
	assertEqualsStr(t, "", "", file)
	assertExpectErr(t, "", err)

}
