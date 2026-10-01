package httptransport

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionCookieSettings(t *testing.T) {
	for _, production := range []bool{false, true} {
		h := authHandler{production: production}
		response := httptest.NewRecorder()
		http.SetCookie(response, h.cookie("token"))
		cookie := response.Result().Cookies()[0]
		wantSameSite := http.SameSiteLaxMode
		if production {
			wantSameSite = http.SameSiteNoneMode
		}
		if cookie.Name != sessionCookie || cookie.Value != "token" || cookie.Path != "/" || !cookie.HttpOnly || cookie.Secure != production || cookie.SameSite != wantSameSite {
			t.Fatalf("production=%v: unexpected cookie %+v", production, cookie)
		}
		response = httptest.NewRecorder()
		h.logout(response, httptest.NewRequest(http.MethodPost, "/auth/logout", nil))
		cookie = response.Result().Cookies()[0]
		if cookie.MaxAge != -1 || cookie.Secure != production || cookie.SameSite != wantSameSite || !cookie.HttpOnly || cookie.Path != "/" {
			t.Fatalf("production=%v: unexpected logout cookie %+v", production, cookie)
		}
	}
}
