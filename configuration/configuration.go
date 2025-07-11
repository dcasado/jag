package configuration

import (
	"flag"
	"os"
)

type Configuration interface {
	ListenAddress() string
	ListenPort() string
	LibraryPath() string
	ThumbnailsPath() string
}

type configuration struct {
	listenAddress  string
	listenPort     string
	libraryPath    string
	thumbnailsPath string
}

func (c configuration) ListenAddress() string {
	return c.listenAddress
}

func (c configuration) ListenPort() string {
	return c.listenPort
}

func (c configuration) LibraryPath() string {
	return c.libraryPath
}

func (c configuration) ThumbnailsPath() string {
	return c.thumbnailsPath
}

func New() Configuration {
	listenAddressEnvVar, exists := os.LookupEnv("LISTEN_ADDRESS")
	if !exists {
		listenAddressEnvVar = "127.0.0.1"
	}
	listenAddress := flag.String("listen-address", listenAddressEnvVar, "Listen address of the service")

	listenPortEnvVar, exists := os.LookupEnv("LISTEN_PORT")
	if !exists {
		listenPortEnvVar = "8080"
	}
	listenPort := flag.String("listen-port", listenPortEnvVar, "Listen port of the service")

	libraryPathEnvVar, exists := os.LookupEnv("LIBRARY_PATH")
	if !exists {
		libraryPathEnvVar = "library"
	}
	libraryPath := flag.String("library-path", libraryPathEnvVar, "Path to the photo library")

	thumbnailsPathEnvVar, exists := os.LookupEnv("THUMBNAILS_PATH")
	if !exists {
		thumbnailsPathEnvVar = ".thumbnails"
	}
	thumbnailsPath := flag.String("thumbnails-path", thumbnailsPathEnvVar, "Path to store the thumbnails")

	flag.Parse()

	return configuration{listenAddress: *listenAddress, listenPort: *listenPort, libraryPath: *libraryPath, thumbnailsPath: *thumbnailsPath}
}
