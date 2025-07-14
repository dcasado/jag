package library

import (
	"log"
	"os"
	"slices"
	"time"
)

type Image struct {
	Date string
	Name string
}

func Images(libraryPath string) []Image {
	var images []Image
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
			images = append(images, Image{Date: file.ModTime().Format(time.DateOnly), Name: file.Name()})
		}
	}

	return images
}
