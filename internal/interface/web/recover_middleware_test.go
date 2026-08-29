package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zaidmasri/business-planning-tool/internal/views"
)

func TestRecover_CatchesPanicAndRendersGenericErrorPage(t *testing.T) {
	templateCache := views.LoadTemplates()

	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom: nil pointer dereference or whatever broke")
	})

	handler := Recover(templateCache)(panicking)

	req := httptest.NewRequest(http.MethodGet, "/plan/does-not-matter", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, genericErrorMessage) {
		t.Errorf("expected body to contain generic error message %q, got: %s", genericErrorMessage, body)
	}
	if strings.Contains(body, "boom") {
		t.Errorf("panic value leaked into response body: %s", body)
	}
}

func TestRecover_DoesNotInterfereWithNormalRequests(t *testing.T) {
	templateCache := views.LoadTemplates()

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fine"))
	})

	handler := Recover(templateCache)(ok)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "fine" {
		t.Errorf("expected body %q, got %q", "fine", rec.Body.String())
	}
}
