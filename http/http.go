package http

import (
	"fmt"
	"log"
	"net/http"

	"davidc.es/jag/configuration"
	"davidc.es/jag/html"
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

	serveMux.HandleFunc("GET /{$}", index(configuration.LibraryPath()))

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
		err := html.Index(w, libraryPath)
		if err != nil {
			log.Printf("error serving index. %v", err)
		}
	}
}
