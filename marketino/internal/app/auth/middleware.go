package auth

import (
	"net/http"
	"strings"

	"github.com/amir1848/marketino/utils/dbutils"
	"github.com/gin-gonic/gin"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("username", claims.Username)
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
		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		authorized, err := userHasResource(claims.Username, resourceKey)
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

		c.Set("username", claims.Username)
		c.Next()
	}
}

func userHasResource(username, resourceKey string) (bool, error) {
	db := dbutils.GetDB()

	var isAdmin bool
	err := db.Raw(`SELECT is_admin FROM users WHERE user_name = ?`, username).Scan(&isAdmin).Error
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
		WHERE ur.user_id = (SELECT id FROM users WHERE user_name = ?) AND r.resource_key = ?
	`, username, resourceKey).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
