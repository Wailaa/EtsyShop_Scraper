package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"EtsyScraper/routes"
)

type MockUserRoute struct {
	isRegisterUserCalled  bool
	isVerifyAccountCalled bool
	isLoginAccountCalled  bool
	isLogOutAccountCalled bool
	isForgotPassReqCalled bool
	isChangePassCalled    bool
	isResetPass           bool
}

func (m *MockUserRoute) RegisterUser(c *gin.Context) {
	m.isRegisterUserCalled = true
}

func (m *MockUserRoute) VerifyAccount(c *gin.Context) {
	m.isVerifyAccountCalled = true
}
func (m *MockUserRoute) LoginAccount(c *gin.Context) {
	m.isLoginAccountCalled = true
}
func (m *MockUserRoute) LogOutAccount(c *gin.Context) {
	m.isLogOutAccountCalled = true
}
func (m *MockUserRoute) ForgotPassReq(c *gin.Context) {
	m.isForgotPassReqCalled = true
}
func (m *MockUserRoute) ChangePass(c *gin.Context) {
	m.isChangePassCalled = true
}
func (m *MockUserRoute) ResetPass(c *gin.Context) {
	m.isResetPass = true
}

func MiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
	}
}

func SecondMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
	}
}

func TestGeneraluserRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	MockedUSer := &MockUserRoute{}
	tests := []struct {
		name     string
		method   string
		path     string
		isCalled func() bool
	}{
		{
			name:     "Check if RegisterUser was called",
			method:   "POST",
			path:     "/auth/register",
			isCalled: func() bool { return MockedUSer.isRegisterUserCalled },
		},
		{
			name:     "Check if VerifyAccount was called",
			method:   "GET",
			path:     "/auth/verifyaccount",
			isCalled: func() bool { return MockedUSer.isVerifyAccountCalled },
		},
		{
			name:     "Check if LoginAccount was called",
			method:   "POST",
			path:     "/auth/login",
			isCalled: func() bool { return MockedUSer.isLoginAccountCalled },
		},
		{
			name:     "Check if LogOutAccount was called",
			method:   "GET",
			path:     "/auth/logout",
			isCalled: func() bool { return MockedUSer.isLogOutAccountCalled },
		},
		{
			name:     "Check if ForgotPassReq was called",
			method:   "POST",
			path:     "/auth/forgotpassword",
			isCalled: func() bool { return MockedUSer.isForgotPassReqCalled },
		},
		{
			name:     "Check if ResetPass was called",
			method:   "POST",
			path:     "/auth/resetpassword",
			isCalled: func() bool { return MockedUSer.isResetPass },
		},
		{
			name:     "Check if ChangePass was called",
			method:   "POST",
			path:     "/auth/changepassword",
			isCalled: func() bool { return MockedUSer.isChangePassCalled },
		},
	}

	User := &routes.UserRoute{UserController: MockedUSer}
	User.GeneraluserRoutes(router, MiddleWare(), SecondMiddleWare())

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.True(t, tc.isCalled())
		})
	}

}
