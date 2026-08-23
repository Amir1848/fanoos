package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/amir1848/fanoos/internal/app/patoken"
	"github.com/amir1848/fanoos/utils/dbutils"
	"github.com/gin-gonic/gin"
)

type resolvedUser struct {
	UserID   int64
	UserName string
}

// resolveUser authenticates a request either by the session cookie set at
// login or by a personal access token in the Authorization header (used
// for API/script access). Session cookies take precedence.
func resolveUser(c *gin.Context) (*resolvedUser, error) {
	if rawToken, err := c.Cookie(sessionCookieName); err == nil && rawToken != "" {
		userID, err := ValidateSession(rawToken)
		if err != nil {
			return nil, err
		}
		return userByID(userID)
	}

	authHeader := c.GetHeader("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString != "" {
		userID, err := patoken.ValidateToken(context.Background(), tokenString)
		if err != nil {
			return nil, err
		}
		return userByID(userID)
	}

	return nil, errInvalidSession
}

func userByID(userID int64) (*resolvedUser, error) {
	var username string
	err := dbutils.GetDB().Raw(`SELECT user_name FROM users WHERE id = ?`, userID).Scan(&username).Error
	if err != nil {
		return nil, err
	}
	if username == "" {
		return nil, errInvalidSession
	}

	return &resolvedUser{UserID: userID, UserName: username}, nil
}

func setUserContext(c *gin.Context, user *resolvedUser) {
	c.Set("username", user.UserName)
	c.Set("userId", user.UserID)
}

// AuthMiddleware authenticates every request in a protected group via the
// session cookie or a Bearer personal access token, and stores the
// resolved user in the context ("userId"/"username").
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := resolveUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}

		setUserContext(c, user)
		c.Next()
	}
}

// RequireResource guards a route behind a resource key. It expects
// AuthMiddleware to have run (protected groups apply it globally) and only
// re-checks the permission.
func RequireResource(resourceKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("userId")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		authorized, err := userHasResource(userID.(int64), resourceKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization check failed"})
			c.Abort()
			return
		}

		if !authorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func userHasResource(userID int64, resourceKey string) (bool, error) {
	db := dbutils.GetDB()

	var isAdmin bool
	err := db.Raw(`SELECT is_admin FROM users WHERE id = ?`, userID).Scan(&isAdmin).Error
	if err != nil {
		return false, err
	}

	if isAdmin {
		return true, nil
	}

	var count int64
	err = db.Raw(`
		SELECT COUNT(1) FROM user_roles ur
		JOIN role_resources rr ON ur.role_id = rr.role_id
		JOIN resources r ON rr.resource_id = r.id
		WHERE ur.user_id = ? AND r.resource_key = ?
	`, userID, resourceKey).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
