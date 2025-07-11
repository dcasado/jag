package library

import (
	"log"
	"os"
	"slices"
)

func ImageNames(libraryPath string) []string {
	var files []string
	f, err := os.Open(libraryPath)
	if err != nil {
		log.Fatalf("error opening library folder. %s. %v", libraryPath, err)
	}

	fileInfos, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Fatalf("error reading the contents of the library folder. %s. %v", libraryPath, err)
	}

	slices.SortFunc(fileInfos, func(a, b os.FileInfo) int { return b.ModTime().Compare(a.ModTime()) })

	for _, file := range fileInfos {
		if !file.IsDir() {
			files = append(files, file.Name())
		}
	}
	return files
}
