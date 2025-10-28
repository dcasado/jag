package http

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

type session struct {
	token               string
	expirationTimestamp time.Time
}

type sessionService interface {
	createSession() (*session, error)
	session(token string) (*session, error)
	deleteSession(token string) error
}

type inMemorySessionService struct {
	sessions             map[string]time.Time
	maxSessionAgeSeconds int
}

func (s inMemorySessionService) createSession() (*session, error) {
	// Make a byte array of size 32
	b := make([]byte, 32)
	// Populate the array with random numbers
	// Ignore the error since it cannot fail. See source for more details
	rand.Read(b)
	// Encode the byte array to an string with hexadecimal encoding
	token := hex.EncodeToString(b)
	expirationTimestamp := time.Now().UTC().Add(time.Duration(s.maxSessionAgeSeconds) * time.Second)
	s.sessions[token] = expirationTimestamp

	// Maintain the error in the interface in case the implementation changes and can return an actual error in the future
	return &session{token: token, expirationTimestamp: expirationTimestamp}, nil
}

func (s inMemorySessionService) session(token string) (*session, error) {
	expirationTime, ok := s.sessions[token]
	if !ok {
		return nil, errors.New("session not found")
	}

	return &session{token: token, expirationTimestamp: expirationTime}, nil
}

func (s inMemorySessionService) deleteSession(token string) error {
	delete(s.sessions, token)
	// Maintain the error in the interface in case the implementation changes and can return an actual error in the future
	return nil
}
