package setupMockServer

import (
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func SetGinTestMode() (*gin.Context, *gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, router := gin.CreateTestContext(w)
	return ctx, router, w
}
