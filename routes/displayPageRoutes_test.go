package routes_test

import (
	"EtsyScraper/routes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGeneralHTMLRoutes(t *testing.T) {

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	testCases := []struct {
		path   string
		status int
	}{
		{"/reset_password", http.StatusOK},
		{"/change_password", http.StatusOK},
		{"/log_in", http.StatusOK},
		{"/verify_account", http.StatusOK},
		{"/stats", http.StatusOK},
		{"/", http.StatusOK},
	}
	templatesFilesPath := "../static/templates/*"
	htmlRoutes := &routes.HTMLRoutes{}
	htmlRoutes.GeneralHTMLRoutes(router, MiddleWare(), SecondMiddleWare(), templatesFilesPath)

	for _, tc := range testCases {

		req, err := http.NewRequest("GET", tc.path, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		router.ServeHTTP(w, req)

		if status := w.Code; status != tc.status {
			t.Errorf("handler returned wrong status code for %s: got %v want %v",
				tc.path, status, tc.status)
		}
	}
}
