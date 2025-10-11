package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

type CSRFStore struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

func NewCSRFStore() *CSRFStore {
	store := &CSRFStore{
		tokens: make(map[string]time.Time),
	}
	go store.cleanupExpired()
	return store
}

func (s *CSRFStore) cleanupExpired() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, expires := range s.tokens {
			if now.After(expires) {
				delete(s.tokens, token)
			}
		}
		s.mu.Unlock()
	}
}

func (s *CSRFStore) Generate() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(bytes)

	s.mu.Lock()
	s.tokens[token] = time.Now().Add(1 * time.Hour)
	s.mu.Unlock()

	return token, nil
}

func (s *CSRFStore) Validate(token string) bool {
	if token == "" {
		return false
	}

	s.mu.RLock()
	expires, exists := s.tokens[token]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(expires) {
		s.mu.Lock()
		delete(s.tokens, token)
		s.mu.Unlock()
		return false
	}

	return true
}

func (s *CSRFStore) Delete(token string) {
	s.mu.Lock()
	delete(s.tokens, token)
	s.mu.Unlock()
}

var GlobalCSRFStore = NewCSRFStore()

func CSRFProtection(next func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		method := e.Request.Method

		if method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH" {
			token := e.Request.FormValue("csrf_token")
			if token == "" {
				token = e.Request.Header.Get("X-CSRF-Token")
			}

			if !GlobalCSRFStore.Validate(token) {
				log.Printf("CSRF validation failed for %s %s", method, e.Request.URL.Path)
				return e.JSON(403, map[string]string{"error": "Invalid or expired CSRF token"})
			}

			GlobalCSRFStore.Delete(token)
		}

		return next(e)
	}
}

func GetCSRFToken(e *core.RequestEvent) (string, error) {
	cookie, err := e.Request.Cookie("csrf_token")
	if err == nil && cookie.Value != "" {
		if GlobalCSRFStore.Validate(cookie.Value) {
			return cookie.Value, nil
		}
	}

	token, err := GlobalCSRFStore.Generate()
	if err != nil {
		return "", err
	}

	e.Response.Header().Add("Set-Cookie", fmt.Sprintf("csrf_token=%s; Path=/; HttpOnly; SameSite=Lax; Max-Age=%d", token, 3600))

	return token, nil
}
