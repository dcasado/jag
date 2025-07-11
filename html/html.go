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

func Index(w io.Writer, libraryPath string) error {
	names := library.ImageNames(libraryPath)
	var data []imageData
	for _, name := range names {
		data = append(data, imageData{ImageName: name, ThumbnailName: library.GetThumbnailName(name)})
	}
	return template.Must(template.New("index.html.tmpl").ParseFS(htmlFiles, "index.html.tmpl")).Execute(w, data)
}
