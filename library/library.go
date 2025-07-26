package library

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
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
	CreationTime  time.Time
	ModTime       time.Time
	Path          string
	Name          string
	ThumbnailPath string
	ThumbnailName string
}

var (
	res    = regexp.MustCompile(`^.*(\d\d\d\d\d\d\d\d_\d\d\d\d\d\d).*$`)
	layout = "20060102_150405"
)

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

	for _, file := range fileInfos {
		if !file.IsDir() {
			imageName := file.Name()
			imagePath := path.Join(year, file.Name())
			images = append(images, Image{
				CreationTime: extractCreationTime(file),
				ModTime:      file.ModTime(), Path: imagePath,
				Name:          imageName,
				ThumbnailName: getThumbnailName(imageName),
				ThumbnailPath: getThumbnailPath(imagePath),
			})
		}
	}

	return images, nil
}

// Try to extract the creation date from the name of the file.
// If that is not possible use file.ModTime() as fallback.
func extractCreationTime(file os.FileInfo) time.Time {
	matches := res.FindStringSubmatch(file.Name())
	if len(matches) >= 2 {
		match := matches[1]
		if len(match) > 0 {
			creationTime, err := time.Parse(layout, match)
			if err != nil {
				fmt.Printf("error extracting creation time from %s. Defaulting to ModTime(). %v\n", file.Name(), err)
				return file.ModTime()
			}
			return creationTime
		}
	}

	fmt.Printf("could not extract creation time from %s. Defaulting to ModTime().\n", file.Name())
	return file.ModTime()
}
