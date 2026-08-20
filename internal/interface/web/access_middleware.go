package web

import (
	"html/template"
	"net/http"

	"uuid"

	ifaces "github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// PlanAccessMiddleware gates routes under /plan/{id}/... by the caller's
// access level on that plan.
type PlanAccessMiddleware struct {
	accessSvc     ifaces.AccessService
	templateCache map[string]*template.Template
}

func NewPlanAccessMiddleware(accessSvc ifaces.AccessService, templateCache map[string]*template.Template) *PlanAccessMiddleware {
	return &PlanAccessMiddleware{accessSvc: accessSvc, templateCache: templateCache}
}

// RequireAccess wraps a handler so it only runs if the logged-in user has at
// least requiredLevel access to the plan identified by the {id} path value.
func (m *PlanAccessMiddleware) RequireAccess(requiredLevel domain.AccessLevel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r)
			if user == nil {
				renderErrorPage(w, r, m.templateCache, http.StatusUnauthorized, "You must be logged in to access this resource.")
				return
			}

			planIDStr := r.PathValue("id")
			if planIDStr == "" {
				// No plan ID in path, allow through (e.g., home page)
				next.ServeHTTP(w, r)
				return
			}

			planID, err := uuid.Parse(planIDStr)
			if err != nil {
				renderErrorPage(w, r, m.templateCache, http.StatusBadRequest, "Invalid plan ID.")
				return
			}

			// Check if user has access to the plan
			access, err := m.accessSvc.GetAccess(planID, user.ID())
			if err != nil {
				renderErrorPage(w, r, m.templateCache, http.StatusForbidden, "You don't have access to this plan.")
				return
			}

			// Check if user has the required access level
			if requiredLevel == domain.Owner && access.AccessLevel != domain.Owner {
				renderErrorPage(w, r, m.templateCache, http.StatusForbidden, "Only the plan owner can perform this action.")
				return
			}
			if requiredLevel == domain.Editor && !access.CanEdit() {
				renderErrorPage(w, r, m.templateCache, http.StatusForbidden, "You don't have permission to edit this plan.")
				return
			}
			if requiredLevel == domain.Viewer && !access.CanView() {
				renderErrorPage(w, r, m.templateCache, http.StatusForbidden, "You don't have permission to view this plan.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
