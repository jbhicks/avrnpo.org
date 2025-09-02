package templates

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/gobuffalo/plush/v4"
	"strings"
)

func TestParseAllTemplates(t *testing.T) {
	if os.Getenv("CI") == "true" || os.Getenv("GO_ENV") == "test" {
		// In CI/test environments we still want to validate templates
	}

	tmplFS := FS()
	err := fs.WalkDir(tmplFS, ".", func(path string, d fs.DirEntry, err error) error {
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
		data, err := fs.ReadFile(tmplFS, path)
		if err != nil {
			return err
		}
		if _, err := plush.Parse(string(data)); err != nil {
			t.Logf("Failed to parse template: %s", path)
			t.Logf("Template content around error:\n%s", string(data))
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Template parse error: %v", err)
	}
}
