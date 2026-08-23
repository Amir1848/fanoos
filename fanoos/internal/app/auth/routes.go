package auth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amir1848/fanoos/utils/dbutils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type getUserFunc func(ctx context.Context, username string) (hashedPassword string, found bool, err error)

var getUserData getUserFunc

var logins = newLoginLimiter()

func RegisterGetUserData(f getUserFunc) {
	getUserData = f
}

func RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/login", func(c *gin.Context) {
		ctx := c.Request.Context()

		var loginData struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&loginData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if ok, retryAfter := logins.allow(c.ClientIP(), loginData.Username); !ok {
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "تلاش‌های ناموفق زیاد؛ کمی بعد دوباره امتحان کنید"})
			return
		}

		storedHash, found, err := getUserData(ctx, loginData.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate"})
			return
		}

		if !found || !CheckPasswordHash(loginData.Password, storedHash) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "نام کاربری یا رمز عبور نادرست است"})
			return
		}

		var row struct {
			ID       int64
			UserName string
		}
		err = dbutils.GetDB().WithContext(ctx).
			Raw(`SELECT id, user_name FROM users WHERE user_name = ?`, loginData.Username).
			Scan(&row).Error
		if err != nil || row.ID == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate"})
			return
		}

		rawToken, err := CreateSession(row.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}

		logins.success(c.ClientIP(), loginData.Username)
		SetSessionCookie(c, rawToken)
		c.JSON(http.StatusOK, gin.H{"username": row.UserName})
	})

	r.POST("/auth/logout", func(c *gin.Context) {
		if rawToken, err := c.Cookie(sessionCookieName); err == nil {
			DeleteSession(rawToken)
		}
		ClearSessionCookie(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}

func RegisterProtectedRoutes(r *gin.RouterGroup) {
	r.GET("/auth/resources", func(c *gin.Context) {
		ctx := c.Request.Context()

		userID, ok := c.Get("userId")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			return
		}

		keys, username, err := UserResourceKeys(ctx, userID.(int64))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load resources"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"username": username, "resources": keys})
	})
}

func CheckPasswordHash(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil // If no error, the password is correct
}
