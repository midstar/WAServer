package main

import (
	"fmt"
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
	res, _ := jsonOfJsons(".test/data/adir")
	fmt.Print(res)
	fmt.Print("\n")
}
