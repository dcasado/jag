package library

import (
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
	imageNames := ImageNames(libraryPath)
	for _, imageName := range imageNames {
		exists, err := Exists(path.Join(thumbnailsPath, imageName))
		if err != nil {
			log.Fatalf("could not check thumbnail existence. %v", err)
		}

		if !exists {
			// Open image file
			imageFile, err := os.Open(path.Join(libraryPath, imageName))
			if err != nil {
				return fmt.Errorf("error opening image file. %w", err)
			}

			_, err = GenerateThumbnail(imageFile, thumbnailsPath)
			imageFile.Close()
			if err != nil {
				log.Printf("could not generate thumbnail for file %s. %v", imageName, err)
			}
		}
	}
	return nil
}

func GenerateThumbnail(imageFile *os.File, thumbnailsPath string) (string, error) {
	// Decode image as jpeg
	inputImage, _, err := image.Decode(imageFile)
	if err != nil {
		return "", fmt.Errorf("error decoding image. %w", err)
	}

	// Extract original size and scale it
	newWidth, newHeight := inputImage.Bounds().Max.X/4, inputImage.Bounds().Max.Y/4
	thumbnailImage := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.ApproxBiLinear.Scale(thumbnailImage, thumbnailImage.Bounds(), inputImage, inputImage.Bounds(), draw.Over, nil)

	// Create thumbnail file
	inputImageInfo, err := os.Lstat(imageFile.Name())
	if err != nil {
		return "", fmt.Errorf("error extracting file info. %w", err)
	}
	thumbnailPath := path.Join(thumbnailsPath, GetThumbnailName(inputImageInfo.Name()))
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

func Exists(imagePath string) (bool, error) {
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

func GetThumbnailName(imageName string) string {
	return strings.TrimSuffix(imageName, filepath.Ext(imageName)) + ".jpg"
}
