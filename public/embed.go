package public

import (
	"embed"
	"github.com/gobuffalo/buffalo"
	"io/fs"
)

//go:embed *
var files embed.FS

func FS() fs.FS {
	return buffalo.NewFS(files, "")
}
