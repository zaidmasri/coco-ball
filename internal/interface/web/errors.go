// Package web - shared error-rendering helpers available to every controller.
package web

import (
	"html/template"
	"net/http"

	"github.com/zaidmasri/business-planning-tool/internal/views"
)

// renderErrorPage renders the shared error page template with the given
// status code and message.
func renderErrorPage(w http.ResponseWriter, r *http.Request, templateCache map[string]*template.Template, statusCode int, message string) {
	page := views.BuildErrorPage(r, statusCode, message)
	views.RenderErrorPageWithStatus(w, templateCache, page, statusCode)
}

// NotFound is the catch-all 404 handler for unmatched URLs.
func NotFound(templateCache map[string]*template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderErrorPage(w, r, templateCache, http.StatusNotFound, "The URL you entered does not exist. Please check the address and try again.")
	}
}
