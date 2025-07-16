package html

import (
	"embed"
	"io"
	"text/template"

	"davidc.es/jag/library"
)

//go:embed *.html.tmpl
var htmlFiles embed.FS

type imageData struct {
	ImageName     string
	ThumbnailName string
}

type bucket struct {
	Date   string
	Images []imageData
}

func Index(w io.Writer, libraryPath string) error {
	images := library.Images(libraryPath)
	var data []*bucket
	for _, image := range images {
		date := image.Date
		if b := containsBucket(data, date); b != nil {
			b.Images = append(b.Images, imageData{ImageName: image.Name, ThumbnailName: library.GetThumbnailName(image.Name)})
		} else {
			newBucket := &bucket{Date: date, Images: make([]imageData, 0)}
			newBucket.Images = append(newBucket.Images, imageData{ImageName: image.Name, ThumbnailName: library.GetThumbnailName(image.Name)})
			data = append(data, newBucket)
		}
	}

	return parseTemplate("index.html.tmpl").Execute(w, data)
}

func containsBucket(buckets []*bucket, date string) *bucket {
	for _, bucket := range buckets {
		if bucket.Date == date {
			return bucket
		}
	}
	return nil
}

func parseTemplate(file string) *template.Template {
	return template.Must(
		template.New("layout.html.tmpl").ParseFS(htmlFiles, "layout.html.tmpl", file))
}
