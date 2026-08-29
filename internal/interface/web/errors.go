// Package web - shared error-rendering helpers available to every controller.
package web

import (
	"html/template"
	"log"
	"net/http"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

// genericErrorMessage is shown to the user whenever an error isn't safe to
// display verbatim (see safeErrorMessage) - it never reveals internals like
// SQL errors, file paths, or Go error strings.
const genericErrorMessage = "Something went wrong on our end. Please try again."

// renderErrorPage renders the shared error page template with the given
// status code and message.
func renderErrorPage(w http.ResponseWriter, r *http.Request, templateCache map[string]*template.Template, statusCode int, message string) {
	page := views.BuildErrorPage(r, statusCode, message)
	views.RenderErrorPageWithStatus(w, templateCache, page, statusCode)
}

// renderInternalError logs err (with context identifying the failing call,
// e.g. "PlanSvc CreatePlan") and renders a generic 500 page. Use this
// instead of surfacing err.Error() directly - the real error may come from
// infrastructure (SQL, filesystem) and should never reach the browser.
func renderInternalError(w http.ResponseWriter, r *http.Request, templateCache map[string]*template.Template, context string, err error) {
	log.Printf("%s: %v", context, err)
	renderErrorPage(w, r, templateCache, http.StatusInternalServerError, genericErrorMessage)
}

// renderCommandError renders the result of a failed write (a command/service
// call): a 400 with the message when err is a known domain sentinel (safe
// to show verbatim - see entities.IsUserFacing), otherwise a generic 500
// with the real error logged server-side.
func renderCommandError(w http.ResponseWriter, r *http.Request, templateCache map[string]*template.Template, context string, err error) {
	if domain.IsUserFacing(err) {
		renderErrorPage(w, r, templateCache, http.StatusBadRequest, err.Error())
		return
	}
	renderInternalError(w, r, templateCache, context, err)
}

// safeErrorMessage returns err's message when it's a known domain sentinel
// (safe to show verbatim - see entities.IsUserFacing), or logs the real
// error server-side with context and returns a generic message otherwise.
// Use this when a handler re-renders a specific form (login/signup/profile)
// with an inline error rather than the shared error page.
func safeErrorMessage(context string, err error) string {
	if domain.IsUserFacing(err) {
		return err.Error()
	}
	log.Printf("%s: %v", context, err)
	return genericErrorMessage
}

// NotFound is the catch-all 404 handler for unmatched URLs.
func NotFound(templateCache map[string]*template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderErrorPage(w, r, templateCache, http.StatusNotFound, "The URL you entered does not exist. Please check the address and try again.")
	}
}
