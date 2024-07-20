package main

import "os"

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
