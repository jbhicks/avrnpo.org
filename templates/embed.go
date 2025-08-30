package templates

import (
	"embed"
	"io/fs"

	"github.com/gobuffalo/buffalo"
)

//go:embed *.plush.html */*.plush.html */*/*.plush.html
var files embed.FS

func FS() fs.FS {
	return buffalo.NewFS(files, "templates")
}
