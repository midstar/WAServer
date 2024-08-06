package main

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestListFilesMap(t *testing.T) {
	// List files in current directory
	m, err := listFilesMap(".")
	assertExpectNoErr(t, "", err)
	assertTrue(t, "", slices.Contains(m["files"], "main.go"))
	assertTrue(t, "", slices.Contains(m["files"], ".gitignore"))
	assertFalse(t, "", slices.Contains(m["files"], ".test"))
	assertFalse(t, "", slices.Contains(m["files"], "app"))
	assertFalse(t, "", slices.Contains(m["dirs"], "main.go"))
	assertFalse(t, "", slices.Contains(m["dirs"], ".gitignore"))
	assertTrue(t, "", slices.Contains(m["dirs"], ".test"))
	assertTrue(t, "", slices.Contains(m["dirs"], "app"))

	// List files in subdirectory
	m, err = listFilesMap(".test/data/adir")
	assertExpectNoErr(t, "", err)
	assertTrue(t, "", slices.Contains(m["files"], "myfile.json"))

	// List files in a non-existing directory
	_, err = listFilesMap("dir/dont/exist")
	assertExpectErr(t, "", err)
}

func TestJsonOfJsons(t *testing.T) {
	// Check a directory without json files
	res, err := jsonOfJsons(".")
	assertExpectNoErr(t, "", err)
	assertEqualsStr(t, "", "{\n}", res)

	// Check a directory with multiple jsons
	res, err = jsonOfJsons(".test/data/adir")
	assertExpectNoErr(t, "", err)
	var m map[string]interface{}
	err = json.Unmarshal([]byte(res), &m)
	assertExpectNoErr(t, "", err)
	_, hasKey := m["myarray"]
	assertTrue(t, "", hasKey)
	_, hasKey = m["myfile"]
	assertTrue(t, "", hasKey)
	_, hasKey = m["myfile2"]
	assertTrue(t, "", hasKey)

	// Check element 2 of myarr
	arr := m["myarray"].([]interface{})
	arr2 := int(arr[2].(float64))
	assertEqualsInt(t, "", 12, arr2)

	// Try a directory that not exist
	_, err = jsonOfJsons("non/existing")
	assertExpectErr(t, "", err)
}
