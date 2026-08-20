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

func resolveUser(tokenString string) (*resolvedUser, error) {
	claims, err := ValidateToken(tokenString)
	if err == nil {
		db := dbutils.GetDB()

		var userID int64
		err := db.Raw(`SELECT id FROM users WHERE user_name = ?`, claims.Username).Scan(&userID).Error
		if err != nil {
			return nil, err
		}

		return &resolvedUser{UserID: userID, UserName: claims.Username}, nil
	}

	userID, err := patoken.ValidateToken(context.Background(), tokenString)
	if err != nil {
		return nil, err
	}

	db := dbutils.GetDB()

	var username string
	err = db.Raw(`SELECT user_name FROM users WHERE id = ?`, userID).Scan(&username).Error
	if err != nil {
		return nil, err
	}

	return &resolvedUser{UserID: userID, UserName: username}, nil
}

func setUserContext(c *gin.Context, user *resolvedUser) {
	c.Set("username", user.UserName)
	c.Set("userId", user.UserID)
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := resolveUser(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		setUserContext(c, user)
		c.Next()
	}
}

func RequireResource(resourceKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := resolveUser(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		authorized, err := userHasResource(user.UserID, resourceKey)
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

		setUserContext(c, user)
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
