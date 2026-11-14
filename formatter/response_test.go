package formatter_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VH288/tc-sdk/formatter"
	"github.com/VH288/tc-sdk/validator"
)

func TestFailMasksUnknownError(t *testing.T) {
	w := httptest.NewRecorder()
	formatter.Fail(w, errors.New("private cause"))
	if w.Code != http.StatusInternalServerError || w.Header().Get("Content-Type") != "application/json" || w.Body.String() != "{\"success\":false,\"error\":{\"code\":\"INTERNAL_ERROR\",\"message\":\"internal server error\"}}\n" {
		t.Fatalf("unexpected response: %d %q %s", w.Code, w.Header().Get("Content-Type"), w.Body.String())
	}
}

func TestValidationFailureResponse(t *testing.T) {
	w := httptest.NewRecorder()
	err := validator.Struct(struct {
		Name string `validate:"required"`
	}{})
	formatter.Fail(w, err)
	want := "{\"success\":false,\"error\":{\"code\":\"VALIDATION_ERROR\",\"message\":\"validation failed\",\"details\":{\"name\":\"failed on 'required' validation\"}}}\n"
	if w.Code != http.StatusUnprocessableEntity || w.Body.String() != want {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestSuccessResponses(t *testing.T) {
	for _, tc := range []struct {
		name  string
		write func(http.ResponseWriter)
		want  string
	}{
		{"success", func(w http.ResponseWriter) { formatter.Success(w, http.StatusCreated, "created", "item") }, "{\"success\":true,\"message\":\"created\",\"data\":\"item\"}\n"},
		{"meta", func(w http.ResponseWriter) {
			formatter.SuccessWithMeta(w, http.StatusCreated, "created", "item", &formatter.Meta{Page: 2, PerPage: 10, Total: 21, TotalPages: 3})
		}, "{\"success\":true,\"message\":\"created\",\"data\":\"item\",\"meta\":{\"page\":2,\"per_page\":10,\"total\":21,\"total_pages\":3}}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tc.write(w)
			if w.Code != http.StatusCreated || w.Header().Get("Content-Type") != "application/json" || w.Body.String() != tc.want {
				t.Fatalf("unexpected response: %d %q %s", w.Code, w.Header().Get("Content-Type"), w.Body.String())
			}
		})
	}
}
