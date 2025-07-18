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
	ImagePath     string
	ThumbnailPath string
}

type bucket struct {
	Date   string
	Images []imageData
}

type indexData struct {
	Years []string
}

func Index(w io.Writer, years []string) error {
	data := indexData{Years: years}

	return parseTemplate("index.html.tmpl").Execute(w, data)
}

func NotFound(w io.Writer) error {
	return parseTemplate("404.html.tmpl").Execute(w, nil)
}

func InternalError(w io.Writer) error {
	return parseTemplate("internal_error.html.tmpl").Execute(w, nil)
}

func Year(w io.Writer, images []library.Image) error {
	var data []*bucket
	for _, image := range images {
		date := image.Date
		if b := containsBucket(data, date); b != nil {
			b.Images = append(b.Images, imageData{ImagePath: image.Path, ThumbnailPath: image.ThumbnailPath})
		} else {
			newBucket := &bucket{Date: date, Images: make([]imageData, 0)}
			newBucket.Images = append(newBucket.Images, imageData{ImagePath: image.Path, ThumbnailPath: image.ThumbnailPath})
			data = append(data, newBucket)
		}
	}

	return parseTemplate("year.html.tmpl").Execute(w, data)
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
