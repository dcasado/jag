package library

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"time"
)

var ErrNotExist = os.ErrNotExist

type ErrUnexpected struct {
	cause error
}

func (e ErrUnexpected) Error() string {
	return fmt.Sprintf("unexpected error. %v", e.cause)
}

type Image struct {
	Date          string
	Path          string
	Name          string
	ThumbnailPath string
	ThumbnailName string
}

func Years(libraryPath string) []string {
	var years []string
	f, err := os.Open(libraryPath)
	if err != nil {
		log.Fatalf("error opening library folder. %s. %v", libraryPath, err)
	}

	fileInfos, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Fatalf("error reading the contents of the library folder. %s. %v", libraryPath, err)
	}

	for _, file := range fileInfos {
		if file.IsDir() {
			years = append(years, file.Name())
		}
	}
	return years
}

func Year(libraryPath string, year string) ([]Image, error) {
	yearPath := path.Join(libraryPath, year)
	_, err := os.Stat(yearPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotExist
		}
		return nil, ErrUnexpected{cause: err}
	}

	var images []Image
	f, err := os.Open(yearPath)
	if err != nil {
		return nil, ErrUnexpected{cause: fmt.Errorf("error opening year folder. %s. %v", yearPath, err)}
	}

	fileInfos, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, ErrUnexpected{cause: fmt.Errorf("error reading the contents of the year folder. %s. %v", yearPath, err)}
	}

	slices.SortFunc(fileInfos, func(a, b os.FileInfo) int { return b.ModTime().Compare(a.ModTime()) })

	for _, file := range fileInfos {
		if !file.IsDir() {
			imageName := file.Name()
			imagePath := path.Join(year, file.Name())
			images = append(images, Image{Date: file.ModTime().Format(time.DateOnly), Path: imagePath, Name: imageName, ThumbnailName: getThumbnailName(imageName), ThumbnailPath: getThumbnailPath(imagePath)})
		}
	}

	return images, nil
}
