package public

import (
	"embed"
	"github.com/gobuffalo/buffalo"
	"io/fs"
)

//go:embed public/assets/**/*
var files embed.FS

func FS() fs.FS {
	return buffalo.NewFS(files, "public/assets")
}
