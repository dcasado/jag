package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"davidc.es/jag/configuration"
	"davidc.es/jag/html"
	"davidc.es/jag/library"
	"davidc.es/jag/static"
)

func Serve(configuration configuration.Configuration) *http.Server {
	serveMux := http.NewServeMux()

	library := http.FileServer(http.Dir(configuration.LibraryPath()))
	serveMux.Handle("GET /library/", http.StripPrefix("/library/", library))
	thumbnails := http.FileServer(http.Dir(configuration.ThumbnailsPath()))
	serveMux.Handle("GET /thumbnails/", http.StripPrefix("/thumbnails/", thumbnails))

	resources := http.FileServerFS(static.Resources())
	serveMux.Handle("GET /resources/", resources)

	serveMux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) { html.NotFound(w) })
	serveMux.HandleFunc("GET /{$}", index(configuration.LibraryPath()))
	serveMux.HandleFunc("GET /{year}", year(configuration.LibraryPath()))

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

func index(libraryPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		years := library.Years(libraryPath)

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

		err = html.Year(w, images)
		if err != nil {
			html.InternalError(w)
			log.Printf("error serving year. %v", err)
			return
		}
	}
}
