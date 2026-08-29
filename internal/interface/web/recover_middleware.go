package web

import (
	"html/template"
	"log"
	"net/http"
	"runtime/debug"
)

// Recover wraps the whole handler chain so a panic in any handler or
// downstream middleware logs its stack trace and renders the shared error
// page instead of the connection dying with no response - net/http's own
// per-request recovery closes the connection without one.
func Recover(templateCache map[string]*template.Template) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("panic handling %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
					renderErrorPage(w, r, templateCache, http.StatusInternalServerError, genericErrorMessage)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
