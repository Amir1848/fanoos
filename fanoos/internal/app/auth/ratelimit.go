package auth

import (
	"sync"
	"time"
)

// loginLimiter is a small in-memory fixed-window rate limiter protecting
// POST /api/login against password guessing. Attempts are tracked per
// client IP and per username (whichever exhausts first blocks the attempt),
// and successful logins clear both counters.
type loginLimiter struct {
	mu       sync.Mutex
	attempts map[string]*window
}

type window struct {
	count    int
	expiresAt time.Time
}

const (
	loginMaxAttempts = 5
	loginWindow      = 15 * time.Minute
)

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{attempts: make(map[string]*window)}
}

// allow reports whether an attempt from ip for username is permitted. If
// not, it also returns how long the caller must wait.
func (l *loginLimiter) allow(ip, username string) (ok bool, retryAfter time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// Lazily purge stale windows so the map can't grow without bound.
	for key, w := range l.attempts {
		if now.After(w.expiresAt) {
			delete(l.attempts, key)
		}
	}

	for _, key := range []string{"ip:" + ip, "user:" + username} {
		if w, found := l.attempts[key]; found && now.Before(w.expiresAt) {
			if w.count >= loginMaxAttempts {
				return false, time.Until(w.expiresAt)
			}
		}
	}

	for _, key := range []string{"ip:" + ip, "user:" + username} {
		w, found := l.attempts[key]
		if !found || now.After(w.expiresAt) {
			w = &window{expiresAt: now.Add(loginWindow)}
			l.attempts[key] = w
		}
		w.count++
	}

	return true, 0
}

// success resets the counters after a successful login.
func (l *loginLimiter) success(ip, username string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.attempts, "ip:"+ip)
	delete(l.attempts, "user:"+username)
}
