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

var templates map[string]*template.Template

func ParseTemplates() {
	templates = make(map[string]*template.Template, 5)
	templates["login"] = template.Must(template.New("login").ParseFS(htmlFiles, "layout.html.tmpl", "login_header.html.tmpl", "login.html.tmpl"))
	templates["index"] = template.Must(template.New("index").ParseFS(htmlFiles, "layout.html.tmpl", "header.html.tmpl", "index.html.tmpl"))
	templates["not_found"] = template.Must(template.New("not_found").ParseFS(htmlFiles, "layout.html.tmpl", "header.html.tmpl", "404.html.tmpl"))
	templates["internal_error"] = template.Must(template.New("internal_error").ParseFS(htmlFiles, "layout.html.tmpl", "header.html.tmpl", "internal_error.html.tmpl"))
	templates["year"] = template.Must(template.New("year").ParseFS(htmlFiles, "layout.html.tmpl", "header.html.tmpl", "year.html.tmpl"))
}

func Login(w io.Writer) error {
	return templates["login"].ExecuteTemplate(w, "base", nil)
}

func Index(w io.Writer, years []string) error {
	data := indexData{Years: years}

	return templates["index"].ExecuteTemplate(w, "base", data)
}

func NotFound(w io.Writer) error {
	return templates["not_found"].ExecuteTemplate(w, "base", nil)
}

func InternalError(w io.Writer) error {
	return templates["internal_error"].ExecuteTemplate(w, "base", nil)
}

func Year(w io.Writer, images []library.Image) error {
	var data []*bucket
	for _, image := range images {
		date := image.CreationTime.Format("January")
		if b := containsBucket(data, date); b != nil {
			b.Images = append(b.Images, imageData{ImagePath: image.Path, ThumbnailPath: image.ThumbnailPath})
		} else {
			newBucket := &bucket{Date: date, Images: make([]imageData, 0)}
			newBucket.Images = append(newBucket.Images, imageData{ImagePath: image.Path, ThumbnailPath: image.ThumbnailPath})
			data = append(data, newBucket)
		}
	}

	return templates["year"].ExecuteTemplate(w, "base", data)
}

func containsBucket(buckets []*bucket, date string) *bucket {
	for _, bucket := range buckets {
		if bucket.Date == date {
			return bucket
		}
	}
	return nil
}
