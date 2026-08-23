package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/amir1848/fanoos/utils/dbutils"
	"github.com/gin-gonic/gin"
)

var errInvalidSession = errors.New("invalid or expired session")

// Session cookie auth (docs/architucture.md §6): the server issues an
// opaque random token in an HttpOnly cookie and stores only its SHA-256
// hash in the sessions table, so a database leak can't be replayed as
// valid sessions.
const (
	sessionCookieName = "fanoos_session"
	sessionLifetime   = 7 * 24 * time.Hour
)

// sessionToken generates a fresh 256-bit session token and its stored hash.
func sessionToken() (rawToken string, tokenHash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}

	rawToken = hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(rawToken))
	return rawToken, hex.EncodeToString(hash[:]), nil
}

// CreateSession persists a new session for the user and returns the raw
// token to hand to the browser. Expired rows are opportunistically purged.
func CreateSession(userID int64) (rawToken string, err error) {
	rawToken, tokenHash, err := sessionToken()
	if err != nil {
		return "", err
	}

	db := dbutils.GetDB()
	if err := db.Exec(`DELETE FROM sessions WHERE expires_at < now()`).Error; err != nil {
		return "", err
	}

	err = db.Exec(
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		tokenHash, userID, time.Now().Add(sessionLifetime),
	).Error
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

// ValidateSession resolves a raw session token to its user, rejecting
// expired sessions. last_used_at is refreshed for observability.
func ValidateSession(rawToken string) (int64, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	db := dbutils.GetDB()

	var userID int64
	err := db.Raw(
		`UPDATE sessions SET last_used_at = now()
		WHERE token_hash = ? AND expires_at > now()
		RETURNING user_id`,
		tokenHash,
	).Scan(&userID).Error
	if err != nil {
		return 0, err
	}
	if userID == 0 {
		return 0, errInvalidSession
	}

	return userID, nil
}

// DeleteSession removes a session (logout).
func DeleteSession(rawToken string) {
	hash := sha256.Sum256([]byte(rawToken))
	dbutils.GetDB().Exec(`DELETE FROM sessions WHERE token_hash = ?`, hex.EncodeToString(hash[:]))
}

// SetSessionCookie writes the session cookie: HttpOnly + SameSite=Lax, so
// JavaScript (and most cross-site requests) can never read or attach it.
// Secure is toggled by SESSION_COOKIE_SECURE for when TLS terminates at
// caddy.
func SetSessionCookie(c *gin.Context, rawToken string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, rawToken, int(sessionLifetime.Seconds()), "/", "", cookieSecure(), true)
}

// ClearSessionCookie expires the session cookie.
func ClearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, "", -1, "/", "", cookieSecure(), true)
}

func cookieSecure() bool {
	v, err := strconv.ParseBool(os.Getenv("SESSION_COOKIE_SECURE"))
	return err == nil && v
}
