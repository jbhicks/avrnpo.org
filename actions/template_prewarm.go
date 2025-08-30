package actions

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gobuffalo/plush/v4"
	"strings"
)

// prewarmTemplates parses all .plush.html templates at startup in dev/test
func prewarmTemplates() error {
	if ENV == "production" {
		return nil
	}

	// Hard-coded absolute templates path to ensure consistent behavior
	tmplDir := "/home/josh/avrnpo.org/templates"
	err := filepath.WalkDir(tmplDir, func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".html" {
			return nil
		}
		if !strings.HasSuffix(path, ".plush.html") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}
		// Try parsing with plush to catch syntax errors
		_, err = plush.Parse(string(data))
		if err != nil {
			return fmt.Errorf("template parse error %s: %w", path, err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
