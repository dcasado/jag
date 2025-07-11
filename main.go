package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"davidc.es/jag/configuration"
	"davidc.es/jag/http"
	"davidc.es/jag/library"
)

func main() {
	configuration := configuration.New()

	err := library.GenerateAllThumbnails(configuration.LibraryPath(), configuration.ThumbnailsPath())
	if err != nil {
		log.Fatalf("error generating thumbnails. %v", err)
	}
	// Generate thumbnails every 1 hour
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			err := library.GenerateAllThumbnails(configuration.LibraryPath(), configuration.ThumbnailsPath())
			log.Fatalf("error generating thumbnails. %v", err)
		}
	}()

	// Attach HTTP handlers to HTTP server
	server := http.Serve(configuration)

	// Handle gracefull shutdown
	errC := make(chan error, 1)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-ctx.Done()

		log.Println("Shutdown signal received")

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer func() {
			stop()
			cancel()
			close(errC)
		}()

		server.SetKeepAlivesEnabled(false)

		if err := server.Shutdown(ctxTimeout); err != nil {
			errC <- err
		}

		log.Println("Shutdown completed")
	}()

	if err := <-errC; err != nil {
		log.Fatalln("error", err)
	}
	log.Print("Exited properly")
}
