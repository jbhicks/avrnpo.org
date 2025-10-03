package services

import (
	"github.com/microcosm-cc/bluemonday"
)

// SanitizeHTML sanitizes user-submitted HTML content to prevent XSS attacks
// while preserving safe formatting elements commonly used in blog posts.
func SanitizeHTML(input string) string {
	// Start with the UGC (User Generated Content) policy which allows common formatting
	p := bluemonday.UGCPolicy()

	// Allow additional attributes for better formatting
	p.AllowAttrs("class").OnElements("p", "div", "span", "code", "pre")
	p.AllowAttrs("href", "title", "target", "rel").OnElements("a")
	p.AllowAttrs("src", "alt", "title", "width", "height").OnElements("img")

	// Allow common block elements that Quill uses
	p.AllowElements("h1", "h2", "h3", "h4", "h5", "h6")
	p.AllowElements("blockquote", "pre", "code")
	p.AllowElements("ul", "ol", "li")
	p.AllowElements("strong", "em", "u", "s", "sub", "sup")
	p.AllowElements("br", "hr")

	// Allow styling attributes that Quill uses (but sanitize the values)
	p.AllowAttrs("style").OnElements("span", "p", "div")

	// Sanitize and return the clean HTML
	return p.Sanitize(input)
}

// SanitizeHTMLStrict provides stricter sanitization for more sensitive contexts
// This removes all HTML and only keeps plain text.
func SanitizeHTMLStrict(input string) string {
	p := bluemonday.StrictPolicy()
	return p.Sanitize(input)
}
