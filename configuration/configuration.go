package configuration

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Configuration interface {
	ListenAddress() string
	ListenPort() string
	SigningKey() string
	EncryptedPassword() string
	MaxSessionAgeSeconds() int
	LibraryPath() string
	ThumbnailsPath() string
}

type configuration struct {
	listenAddress        string
	listenPort           string
	signingKey           string
	encryptedPassword    string
	maxSessionAgeSeconds int
	libraryPath          string
	thumbnailsPath       string
}

func (c configuration) ListenAddress() string {
	return c.listenAddress
}

func (c configuration) ListenPort() string {
	return c.listenPort
}

func (c configuration) SigningKey() string {
	return c.signingKey
}

func (c configuration) EncryptedPassword() string {
	return c.encryptedPassword
}

func (c configuration) MaxSessionAgeSeconds() int {
	return c.maxSessionAgeSeconds
}

func (c configuration) LibraryPath() string {
	return c.libraryPath
}

func (c configuration) ThumbnailsPath() string {
	return c.thumbnailsPath
}

func New() (Configuration, error) {
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

	signingKeyEnvVar, exists := os.LookupEnv("SIGNING_KEY")
	if !exists {
		// Make a byte array of size 32
		b := make([]byte, 32)
		// Populate the array with random numbers
		// Ignore the error since it cannot fail. See source for more details
		rand.Read(b)
		signingKeyEnvVar = string(b)
	}
	signingKey := flag.String("signing-key", signingKeyEnvVar, "Signing key to use for the session")

	encryptedPasswordEnvVar, exists := os.LookupEnv("ENCRYPTED_PASSWORD")
	if !exists {
		encryptedPasswordEnvVar = ""
	}
	encryptedPassword := flag.String("encrypted-password", encryptedPasswordEnvVar, "bcrypt encrypted password")

	maxSessionAgeSecondsEnvVarStr, exists := os.LookupEnv("MAX_SESSION_AGE_SECONDS")
	if !exists {
		maxSessionAgeSecondsEnvVarStr = "300" // 5 minute default
	}
	maxSessionAgeSecondsEnvVar, err := strconv.Atoi(maxSessionAgeSecondsEnvVarStr)
	if err != nil {
		return nil, fmt.Errorf("MAX_SESSION_AGE_SECONDS must be a number. %w", err)
	}
	maxSessionAgeSeconds := flag.Int("max-session-age-seconds", maxSessionAgeSecondsEnvVar, "Max time in seconds the session will be valid for")

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

	if len(*encryptedPassword) == 0 {
		return nil, errors.New("encrypted password is mandatory and must not be empty")
	}

	return configuration{
		listenAddress:        *listenAddress,
		listenPort:           *listenPort,
		signingKey:           *signingKey,
		encryptedPassword:    *encryptedPassword,
		maxSessionAgeSeconds: *maxSessionAgeSeconds,
		libraryPath:          *libraryPath,
		thumbnailsPath:       *thumbnailsPath,
	}, nil
}
