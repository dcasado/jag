package library

import (
	"embed"
	"errors"
	"fmt"
	"image"
	jpeg "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

const videoThumbnailName string = "video-thumbnail.jpg"

//go:embed thumbnails/*
var thumbnails embed.FS

func GenerateAllThumbnails(libraryPath string, thumbnailsPath string) error {
	_, err := os.Stat(thumbnailsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(thumbnailsPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating thumbnails directory. %w", err)
			}
		} else {
			return fmt.Errorf("error checking if thumbnails directory %s exsits. %w", thumbnailsPath, err)
		}
	}

	err = copyVideoThumbnail(thumbnailsPath)
	if err != nil {
		return fmt.Errorf("error copying video thumbnail to thumbnails folder. %w", err)
	}

	years := Years(libraryPath)
	for _, year := range years {
		libraryYearPath := path.Join(libraryPath, year)
		thumbnailYearPath := path.Join(thumbnailsPath, year)

		err := os.Mkdir(thumbnailYearPath, os.ModePerm)
		if err != nil {
			if !errors.Is(err, os.ErrExist) {
				return fmt.Errorf("error creating thumbnails directory. %w", err)
			}
		}

		images, err := Year(libraryPath, year)
		if err != nil {
			return fmt.Errorf("error retrieving year images. %v", err)
		}
		for _, image := range images {
			if isVideo(image.Name) {
				// Do not try to generate thumbnail for videos
				continue
			}
			exists, err := exists(path.Join(thumbnailYearPath, image.Name))
			if err != nil {
				log.Fatalf("could not check thumbnail existence. %v", err)
			}

			if !exists {
				// Open image file
				imageFile, err := os.Open(path.Join(libraryYearPath, image.Name))
				if err != nil {
					return fmt.Errorf("error opening image file. %w", err)
				}

				_, err = generateThumbnail(imageFile, thumbnailYearPath)
				imageFile.Close()
				if err != nil {
					log.Printf("could not generate thumbnail for file %s. %v", image.Name, err)
				}
			}
		}
	}
	return nil
}

func copyVideoThumbnail(thumbnailsPath string) error {
	videoThumbnail, err := thumbnails.ReadFile(path.Join("thumbnails", videoThumbnailName))
	if err != nil {
		return fmt.Errorf("error reading embed video thumbnail. %w", err)
	}
	err = os.WriteFile(path.Join(thumbnailsPath, videoThumbnailName), videoThumbnail, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error saving video thumbnail. %w", err)
	}
	return nil
}

func generateThumbnail(imageFile *os.File, thumbnailsPath string) (string, error) {
	// Decode image as jpeg
	inputImage, _, err := image.Decode(imageFile)
	if err != nil {
		return "", fmt.Errorf("error decoding image. %w", err)
	}

	// Extract ratio and scale the image
	ratio := float64(inputImage.Bounds().Max.X) / float64(inputImage.Bounds().Bounds().Max.Y)
	newWidth, newHeight := 350, int(float64(350)/ratio)
	thumbnailImage := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.ApproxBiLinear.Scale(thumbnailImage, thumbnailImage.Bounds(), inputImage, inputImage.Bounds(), draw.Over, nil)

	// Create thumbnail file
	inputImageInfo, err := os.Lstat(imageFile.Name())
	if err != nil {
		return "", fmt.Errorf("error extracting file info. %w", err)
	}
	thumbnailPath := path.Join(thumbnailsPath, getThumbnailName(inputImageInfo.Name()))
	thumbnailFile, err := os.Create(thumbnailPath)
	if err != nil {
		return "", fmt.Errorf("error creating thumbnail file %s. %w", thumbnailPath, err)
	}
	defer thumbnailFile.Close()

	// Write thumbnail data to file
	err = jpeg.Encode(thumbnailFile, thumbnailImage, &jpeg.Options{Quality: 70})
	if err != nil {
		return "", fmt.Errorf("error encoding thumbnail image. %w", err)
	}
	return path.Join(thumbnailsPath, thumbnailFile.Name()), nil
}

func exists(imagePath string) (bool, error) {
	_, err := os.Stat(imagePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		} else {
			return false, fmt.Errorf("error checking if file %s exsits. %w", imagePath, err)
		}
	}
	return true, nil
}

func getThumbnailName(imageName string) string {
	if isVideo(imageName) {
		return videoThumbnailName
	}
	return strings.TrimSuffix(imageName, filepath.Ext(imageName)) + ".jpg"
}

func getThumbnailPath(imagePath string) string {
	if isVideo(imagePath) {
		return videoThumbnailName
	}
	return strings.TrimSuffix(imagePath, filepath.Ext(imagePath)) + ".jpg"
}

func isVideo(path string) bool {
	return filepath.Ext(path) == ".mp4"
}
