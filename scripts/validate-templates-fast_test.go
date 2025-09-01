package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempTemplate(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func Test_lintPlushEmittingControl_flags_non_emitting_wrapping_html(t *testing.T) {
	d := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(d)
	// create templates tree
	tmpl := `
<% if (true) { %>
<div>Hi</div>
<% } %>
`
	writeTempTemplate(t, d, "templates/example.bad.plush.html", tmpl)

	findings := lintPlushEmittingControl(false)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %v", len(findings), findings)
	}
	if !strings.Contains(findings[0], "example.bad.plush.html") {
		t.Fatalf("finding should mention file: %v", findings[0])
	}
}

func Test_lintPlushEmittingControl_allows_emitting_blocks(t *testing.T) {
	d := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(d)
	tmpl := `
<%= if (true) { %>
<div>Hi</div>
<% } %>
`
	writeTempTemplate(t, d, "templates/example.good.plush.html", tmpl)

	findings := lintPlushEmittingControl(false)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %v", len(findings), findings)
	}
}

func Test_lintPlushEmittingControl_ignores_script_and_style(t *testing.T) {
	d := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(d)
	tmpl := `
<style>
  <% if (true) { %>
  .x { color: red; }
  <% } %>
</style>
<script>
  <% if (true) { %>
  console.log('x')
  <% } %>
</script>
`
	writeTempTemplate(t, d, "templates/example.style.plush.html", tmpl)

	findings := lintPlushEmittingControl(false)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %v", len(findings), findings)
	}
}

func Test_lintPlushEmittingControl_allowlist_comment(t *testing.T) {
	d := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(d)
	tmpl := `
<% if (true) { %> <%# validator:emit-ok %>
<div>Intentionally suppressed</div>
<% } %>
`
	writeTempTemplate(t, d, "templates/example.allow.plush.html", tmpl)

	findings := lintPlushEmittingControl(false)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %v", len(findings), findings)
	}
}
