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
	"path"
	"slices"
	"strings"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"
const dataPath = ".test/data"

func startServer(t *testing.T) {
	t.Helper()

	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"test", "app", dataPath}
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

func postObject(t *testing.T, path string, expectedStatus int,
	recv_obj interface{}, send_obj interface{}) {
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

func deleteObject(t *testing.T, path string, expectedStatus int, recv_obj interface{}) {
	t.Helper()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", baseURL, path), bytes.NewBuffer(nil))
	if err != nil {
		t.Fatalf("Unable create req for delete path %s. Reason: %s", path, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Unable to delete path %s. Reason: %s", path, err)
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

func TestStaticGet(t *testing.T) {
	startServer(t)
	defer shutdownServer(t)

	html := getHTML(t, "app/3_in_a_row/index.html")
	if !strings.Contains(html, "<title>3 in a row</title>") {
		t.Fatal("Index html title missing")
	}
}

func TestDataDelete(t *testing.T) {
	startServer(t)
	defer shutdownServer(t)

	// Try to delete root directory (=forbidden)
	var recv_obj map[string]string
	deleteObject(t, "data/", http.StatusForbidden, &recv_obj)

	// Tro to delete a directory that dont exist
	deleteObject(t, "data/directory/dont/exist/", http.StatusNotFound, &recv_obj)

	// Tro to delete a file that dont exist
	deleteObject(t, "data/adir/filedontexist", http.StatusNotFound, &recv_obj)

	// Create a directory with three files
	tmpPath := path.Join(dataPath, "deleteTest")
	os.Remove(tmpPath)
	send_obj := map[string]int{"id": 1}

	postObject(t, "data/deleteTest/one", http.StatusOK, &recv_obj, &send_obj)
	assertFileExist(t, "", path.Join(tmpPath, "one.json"))

	send_obj["id"] = 2
	postObject(t, "data/deleteTest/two", http.StatusOK, &recv_obj, &send_obj)
	assertFileExist(t, "", path.Join(tmpPath, "two.json"))

	send_obj["id"] = 3
	postObject(t, "data/deleteTest/three", http.StatusOK, &recv_obj, &send_obj)
	assertFileExist(t, "", path.Join(tmpPath, "three.json"))

	// Delete three.js
	deleteObject(t, "data/deleteTest/two", http.StatusOK, &recv_obj)
	assertFileExist(t, "", tmpPath)
	assertFileExist(t, "", path.Join(tmpPath, "one.json"))
	assertFileNotExist(t, "", path.Join(tmpPath, "two.json"))
	assertFileExist(t, "", path.Join(tmpPath, "three.json"))

	// Delete the whole path
	deleteObject(t, "data/deleteTest/", http.StatusOK, &recv_obj)
	assertFileNotExist(t, "", tmpPath)

}

func TestDataPost(t *testing.T) {
	startServer(t)
	defer shutdownServer(t)

	// Post object
	filePath := path.Join(dataPath, "myjson.json")
	os.Remove(filePath)
	send_obj := map[string]int{"foo": 1, "bar": 2}
	var recv_obj map[string]string
	postObject(t, "data/myjson", http.StatusOK, &recv_obj, &send_obj)
	_, hasMessageKey := recv_obj["message"]
	assertTrue(t, "Key: message not defined", hasMessageKey)
	assertEqualsStr(t, "invalid message", "JSON post successfull", recv_obj["message"])
	assertFileExist(t, "", filePath)

	// Receive object
	var recv_obj2 map[string]int
	getObject(t, "data/myjson", http.StatusOK, &recv_obj2)
	assertEqualsInt(t, "", 1, recv_obj2["foo"])

	// Update object
	send_obj["foo"] = 5
	postObject(t, "data/myjson", http.StatusOK, &recv_obj, &send_obj)

	// Receive updated object
	getObject(t, "data/myjson", http.StatusOK, &recv_obj2)
	assertEqualsInt(t, "", 5, recv_obj2["foo"])

	// Post into subdirectories
	filePath = path.Join(dataPath, "a/deep/dir/structure", "myjson2.json")
	os.Remove(filePath)
	send_obj["foo"] = 9
	postObject(t, "data/a/deep/dir/structure/myjson2", http.StatusOK, &recv_obj, &send_obj)
	assertFileExist(t, "", filePath)

	// Post directory (not allowed)
	postObject(t, "data/directory/", http.StatusForbidden, &recv_obj, &send_obj)

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

	// GET directory - ls query
	var filesMap map[string][]string
	getObject(t, "data/adir/?ls=true", http.StatusOK, &filesMap)
	assertTrue(t, "", slices.Contains(filesMap["files"], "myfile.json"))

	// GET directory - ls query - not found
	var resp map[string]string
	getObject(t, "data/this/dir/dont/exist/?ls=true", http.StatusNotFound, &resp)

	// GET directory
	var m map[string]interface{}
	getObject(t, "data/adir/", http.StatusOK, &m)
	_, hasKey := m["myarray"]
	assertTrue(t, "", hasKey)
	// Check element 2 of myarr
	arr := m["myarray"].([]interface{})
	arr2 := int(arr[2].(float64))
	assertEqualsInt(t, "", 12, arr2)

	// GET directory - not found
	getObject(t, "data/this/dir/dont/exist/", http.StatusNotFound, &resp)

}

func TestTLS(t *testing.T) {
	var baseHttpsURL = "https://localhost:8080"

	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"test", "-d", "-s", "-c", ".test/cert.pem",
		"-k", ".test/key.pem", "app", dataPath}
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
