package httptransport

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/trade-diary/backend/internal/domain/auth"
)

const sessionCookie = "trade_diary_session"

type authHandler struct {
	service    *auth.Service
	production bool
}

func (h authHandler) cookie(value string) *http.Cookie {
	sameSite := http.SameSiteLaxMode
	if h.production {
		sameSite = http.SameSiteNoneMode
	}
	return &http.Cookie{Name: sessionCookie, Value: value, Path: "/", HttpOnly: true, Secure: h.production, SameSite: sameSite}
}

type credentials struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h authHandler) register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decode(w, r, &in) {
		return
	}
	u, err := h.service.Register(r.Context(), in.Name, in.Email, in.Password)
	if err != nil {
		authError(w, r, err)
		return
	}
	if !h.startSession(w, r, u) {
		return
	}
	writeJSON(w, http.StatusCreated, u)
}
func (h authHandler) login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decode(w, r, &in) {
		return
	}
	u, err := h.service.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		authError(w, r, err)
		return
	}
	if !h.startSession(w, r, u) {
		return
	}
	writeJSON(w, http.StatusOK, u)
}
func (h authHandler) me(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.service)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, u)
}
func (h authHandler) logout(w http.ResponseWriter, r *http.Request) {
	if token, ok := cookieHash(r); ok {
		if err := h.service.Logout(r.Context(), token); err != nil {
			internalError(w, r, "auth.logout", err)
			return
		}
	}
	cookie := h.cookie("")
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusNoContent)
}
func (h authHandler) startSession(w http.ResponseWriter, r *http.Request, u auth.User) bool {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		internalError(w, r, "auth.session_token", err, "user_id", u.ID)
		return false
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	expires := time.Now().Add(30 * 24 * time.Hour)
	if err := h.service.CreateSession(r.Context(), u.ID, hashToken(token), expires); err != nil {
		internalError(w, r, "auth.create_session", err, "user_id", u.ID)
		return false
	}
	cookie := h.cookie(token)
	cookie.Expires = expires
	http.SetCookie(w, cookie)
	return true
}
func cookieHash(r *http.Request) (string, bool) {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return "", false
	}
	return hashToken(c.Value), true
}
func authenticatedUser(w http.ResponseWriter, r *http.Request, service *auth.Service) (auth.User, bool) {
	token, ok := cookieHash(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return auth.User{}, false
	}
	user, err := service.CurrentUser(r.Context(), token)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return auth.User{}, false
	}
	if err != nil {
		internalError(w, r, "auth.current_user", err)
		return auth.User{}, false
	}
	return user, true
}
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if decoder.Decode(v) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		writeError(w, 400, "invalid request")
		return false
	}
	return true
}
func authError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, auth.ErrEmailTaken) {
		writeError(w, 409, "email already registered")
		return
	}
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, 400, "invalid credentials")
		return
	}
	internalError(w, r, "auth.authenticate", err)
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
