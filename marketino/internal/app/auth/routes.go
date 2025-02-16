package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type getUserFunc func(ctx context.Context, username string) (hashedPassword string, found bool, err error)

var getUserData getUserFunc

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

		storedHash, found, err := getUserData(ctx, loginData.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		if !found {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		if CheckPasswordHash(loginData.Password, storedHash) {
			token, err := GenerateToken(loginData.Username)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"token": token})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		}
	})
}

func CheckPasswordHash(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil // If no error, the password is correct
}
