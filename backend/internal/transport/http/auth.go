package httptransport

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/trade-diary/backend/internal/domain/auth"
)

const sessionCookie = "trade_diary_session"

type authHandler struct{ service *auth.Service }
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
		authError(w, err)
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
		authError(w, err)
		return
	}
	if !h.startSession(w, r, u) {
		return
	}
	writeJSON(w, http.StatusOK, u)
}
func (h authHandler) me(w http.ResponseWriter, r *http.Request) {
	token, ok := cookieHash(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	u, err := h.service.CurrentUser(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	writeJSON(w, http.StatusOK, u)
}
func (h authHandler) logout(w http.ResponseWriter, r *http.Request) {
	if token, ok := cookieHash(r); ok {
		_ = h.service.Logout(r.Context(), token)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}
func (h authHandler) startSession(w http.ResponseWriter, r *http.Request, u auth.User) bool {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		writeError(w, 500, "internal error")
		return false
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	expires := time.Now().Add(30 * 24 * time.Hour)
	if err := h.service.CreateSession(r.Context(), u.ID, hashToken(token), expires); err != nil {
		writeError(w, 500, "internal error")
		return false
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", Expires: expires, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	return true
}
func cookieHash(r *http.Request) (string, bool) {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return "", false
	}
	return hashToken(c.Value), true
}
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if json.NewDecoder(r.Body).Decode(v) != nil {
		writeError(w, 400, "invalid request")
		return false
	}
	return true
}
func authError(w http.ResponseWriter, err error) {
	if errors.Is(err, auth.ErrEmailTaken) {
		writeError(w, 409, "email already registered")
		return
	}
	writeError(w, 400, "invalid credentials")
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
