package static

import (
	"embed"
	"io/fs"
)

//go:embed resources/*
var files embed.FS

func Resources() fs.FS {
	return files
}
