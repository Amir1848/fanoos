package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amir1848/fanoos/utils/dbutils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// TestMain wires the getUserData callback that the users package normally
// registers at init time (importing it here would be a cycle).
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	RegisterGetUserData(func(ctx context.Context, username string) (string, bool, error) {
		var hash string
		err := dbutils.GetDB().WithContext(ctx).
			Raw(`SELECT password_c FROM users WHERE user_name = ?`, username).Scan(&hash).Error
		return hash, hash != "", err
	})

	m.Run()
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	RegisterRoutes(api)
	protected := api.Group("")
	protected.Use(AuthMiddleware())
	RegisterProtectedRoutes(protected)
	return r
}

func loginRequest(t *testing.T, r *gin.Engine, username, password string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(map[string]string{"username": username, "password": password})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestLoginSetsSessionCookie(t *testing.T) {
	logins = newLoginLimiter()
	r := newTestRouter()

	rec := loginRequest(t, r, "admin", "admin")
	require.Equal(t, http.StatusOK, rec.Code)

	var body struct{ Username string }
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "admin", body.Username)

	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			found = true
			require.NotEmpty(t, c.Value)
			require.True(t, c.HttpOnly)
			require.Equal(t, http.SameSiteLaxMode, c.SameSite)
		}
	}
	require.True(t, found, "session cookie not set")
}

func TestLoginWrongPassword(t *testing.T) {
	logins = newLoginLimiter()
	r := newTestRouter()

	rec := loginRequest(t, r, "admin", "wrong")
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLoginRateLimited(t *testing.T) {
	logins = newLoginLimiter()
	r := newTestRouter()

	for i := 0; i < 5; i++ {
		rec := loginRequest(t, r, "admin", "wrong")
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	}

	rec := loginRequest(t, r, "admin", "wrong")
	require.Equal(t, http.StatusTooManyRequests, rec.Code)
	require.NotEmpty(t, rec.Header().Get("Retry-After"))

	// Even the correct password is blocked while the window is open.
	rec = loginRequest(t, r, "admin", "admin")
	require.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestSessionCookieAuthorizesProtectedRoute(t *testing.T) {
	logins = newLoginLimiter()
	r := newTestRouter()

	rec := loginRequest(t, r, "admin", "admin")
	require.Equal(t, http.StatusOK, rec.Code)
	sessionCookie := rec.Result().Cookies()[0]

	req := httptest.NewRequest(http.MethodGet, "/api/auth/resources", nil)
	req.AddCookie(sessionCookie)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req)

	require.Equal(t, http.StatusOK, rec2.Code)

	var body struct {
		Username  string
		Resources []string
	}
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &body))
	require.Equal(t, "admin", body.Username)
}

func TestProtectedRouteWithoutSession(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/resources", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLogoutInvalidatesSession(t *testing.T) {
	logins = newLoginLimiter()
	r := newTestRouter()

	rec := loginRequest(t, r, "admin", "admin")
	require.Equal(t, http.StatusOK, rec.Code)
	sessionCookie := rec.Result().Cookies()[0]

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(sessionCookie)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req)
	require.Equal(t, http.StatusOK, rec2.Code)

	// The old cookie no longer authorizes anything.
	req = httptest.NewRequest(http.MethodGet, "/api/auth/resources", nil)
	req.AddCookie(sessionCookie)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req)
	require.Equal(t, http.StatusUnauthorized, rec3.Code)
}
