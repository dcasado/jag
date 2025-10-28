package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"davidc.es/jag/configuration"
	"davidc.es/jag/html"
	"davidc.es/jag/library"
	"davidc.es/jag/static"
	"golang.org/x/crypto/bcrypt"
)

const cookieName string = "session"

func Serve(configuration configuration.Configuration) *http.Server {
	sessionService := inMemorySessionService{
		sessions:             make(map[string]time.Time),
		maxSessionAgeSeconds: configuration.MaxSessionAgeSeconds(),
	}

	serveMux := http.NewServeMux()

	serveMux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) { html.Login(w) })
	serveMux.HandleFunc("POST /login", login(configuration.SigningKey(), configuration.EncryptedPassword(), configuration.MaxSessionAgeSeconds(), sessionService))
	serveMux.HandleFunc("POST /logout", auth(configuration.SigningKey(), sessionService, logout(sessionService)))

	library := http.FileServer(http.Dir(configuration.LibraryPath()))
	serveMux.HandleFunc("GET /library/", auth(configuration.SigningKey(), sessionService, http.StripPrefix("/library/", library).ServeHTTP))
	thumbnails := http.FileServer(http.Dir(configuration.ThumbnailsPath()))
	serveMux.Handle("GET /thumbnails/", auth(configuration.SigningKey(), sessionService, http.StripPrefix("/thumbnails/", thumbnails).ServeHTTP))

	resources := http.FileServerFS(static.Resources())
	serveMux.Handle("GET /resources/", resources)

	serveMux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) { html.NotFound(w) })
	serveMux.HandleFunc("GET /{$}", auth(configuration.SigningKey(), sessionService, index(configuration.LibraryPath())))
	serveMux.HandleFunc("GET /{year}", auth(configuration.SigningKey(), sessionService, year(configuration.LibraryPath())))

	serveMux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("Ok"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", configuration.ListenAddress(), configuration.ListenPort()),
		Handler: serveMux,
	}

	// Handle gracefull shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting the server. %s", err)
		}
	}()
	log.Printf("Started server listening on %s:%s", configuration.ListenAddress(), configuration.ListenPort())

	return server
}

func login(signingKey string, encryptedPassword string, maxSessionAgeSeconds int, sessionService sessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		formPassword := r.FormValue("password")
		err = bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(formPassword))
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		} else {
			session, err := sessionService.createSession()
			if err != nil {
				log.Printf("unable to create session: %v", err)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			token := session.token
			// Calculate a HMAC signature of the cookie name and value, using SHA256 and a secret key.
			mac := hmac.New(sha256.New, []byte(signingKey))
			mac.Write([]byte(cookieName))
			mac.Write([]byte(token))
			signature := mac.Sum(nil)

			// Prepend the token with the HMAC signature and encode it to base64.
			cookieValue := base64.URLEncoding.EncodeToString([]byte(string(signature) + token))

			cookie := &http.Cookie{
				Name:     cookieName,
				Value:    cookieValue,
				MaxAge:   maxSessionAgeSeconds,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				HttpOnly: true,
			}
			http.SetCookie(w, cookie)

			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

func auth(signingKey string, sessionService sessionService, authenticatedHandlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		signedToken, err := base64.URLEncoding.DecodeString(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Check that the signed token is at least the size of the signature
		if len(signedToken) < sha256.Size {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Split apart the signature and the token.
		signature := signedToken[:sha256.Size]
		token := signedToken[sha256.Size:]

		// Recalculate the HMAC signature of the cookie name and the token.
		mac := hmac.New(sha256.New, []byte(signingKey))
		mac.Write([]byte(cookieName))
		mac.Write([]byte(token))
		expectedSignature := mac.Sum(nil)

		if !hmac.Equal(signature, expectedSignature) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, err := sessionService.session(string(token))
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if string(token) != session.token || session.expirationTimestamp.Before(time.Now().UTC()) {
			err = sessionService.deleteSession(string(token))
			if err != nil {
				log.Printf("could not delete session: %v", err)
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		authenticatedHandlerFunc(w, r)
	}
}

func index(libraryPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		years := library.Years(libraryPath)

		// Sort years in descending natural sort order
		slices.SortFunc(years, func(a, b string) int { return strings.Compare(b, a) })

		err := html.Index(w, years)
		if err != nil {
			html.InternalError(w)
			log.Printf("error serving index. %v", err)
			return
		}
	}
}

func year(libraryPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		year := r.PathValue("year")

		images, err := library.Year(libraryPath, year)
		if err != nil {
			if errors.Is(err, library.ErrNotExist) {
				html.NotFound(w)
				return
			}
			html.InternalError(w)
			return
		}

		slices.SortFunc(images, func(a, b library.Image) int { return b.ModTime.Compare(a.ModTime) })

		err = html.Year(w, images)
		if err != nil {
			html.InternalError(w)
			log.Printf("error serving year. %v", err)
			return
		}
	}
}

func logout(sessionService sessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		signedToken, err := base64.URLEncoding.DecodeString(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Check that the signed token is at least the size of the signature
		if len(signedToken) < sha256.Size {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Split apart the signature and the token
		token := signedToken[sha256.Size:]
		err = sessionService.deleteSession(string(token))
		if err != nil {
			log.Printf("could not delete session. %v", err)
		}

		cookie = &http.Cookie{
			Name:   cookieName,
			MaxAge: -1,
		}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
