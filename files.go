package main

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// Lists files into a map of following structure:
//
//	{
//	  "files" : ["file1", "file2", ...]
//	  "dirs" : ["dir1", "dir2", ...]
//	}
func listFilesMap(dir string) (map[string][]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	result := map[string][]string{
		"files": {},
		"dirs":  {},
	}
	for _, file := range files {
		if file.IsDir() {
			result["dirs"] = append(result["dirs"], file.Name())
		} else {
			result["files"] = append(result["files"], file.Name())
		}
	}
	return result, nil
}

// Gets all .json files and creates a new json including all
// .json files. For example
//
// filea.json = {"a" : 1, "b" : 2}
// fileb.json = [1,2,3,4]
//
// Will result in:
//
//	{
//	  "filea" : {"a" : 1, "b" : 2},
//	  "fileb" : [1,2,3,4]
//	}
func jsonOfJsons(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	isFirst := true // Flag for , between key : values
	result.WriteString("{")
	for _, file := range files {
		if !file.IsDir() && path.Ext(file.Name()) == ".json" {
			if !isFirst {
				result.WriteString(",") // Add separator
			}
			result.WriteString("\n")
			// Write key (file name without extension)
			name := strings.TrimSuffix(file.Name(), ".json")
			result.WriteString(fmt.Sprintf(`"%s":`, name))
			// Write value (file contents)
			fullPath := path.Join(dir, file.Name())
			dat, _ := os.ReadFile(fullPath)
			result.Write(dat)
			isFirst = false
		}
	}
	result.WriteString("\n}")
	return result.String(), nil
}
