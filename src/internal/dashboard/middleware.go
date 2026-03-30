package dashboard

import (
	"context"
	"net/http"
)

type ctxKey int

const ctxKeyUsername ctxKey = 1

func withRequestUser(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, ctxKeyUsername, username)
}

func requestUser(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(ctxKeyUsername).(string)
	return u, ok && u != ""
}

const sessionCookieName = "optitrade_session"

func sessionCookieSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookieName)
		if err != nil || c == nil || c.Value == "" {
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "sign in required")
			return
		}
		user, err := s.sessions.LookupUsername(r.Context(), c.Value)
		if err != nil {
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "sign in required")
			return
		}
		s.authMu.RLock()
		auth := s.auth
		s.authMu.RUnlock()
		if auth == nil || auth.Lookup(user) == nil {
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "session no longer valid")
			return
		}
		next.ServeHTTP(w, r.WithContext(withRequestUser(r.Context(), user)))
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   sessionCookieSecure(r),
	})
}

func (s *Server) setSessionCookie(w http.ResponseWriter, r *http.Request, rawToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    rawToken,
		Path:     "/",
		MaxAge:   86400 * 365 * 10,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   sessionCookieSecure(r),
	})
}
