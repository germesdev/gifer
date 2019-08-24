package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionHandler(t *testing.T) {
	t.Run("response OK", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/version", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		versionHandler().ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
		}
	})
}
