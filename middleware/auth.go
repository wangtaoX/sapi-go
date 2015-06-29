package middleware

import (
	"encoding/base64"
	"github.com/codegangsta/negroni"
	"net/http"
	"strings"
)

var (
	admitUsers = map[string]string{
		"admin":  "admin",
		"sinanp": "sinanp",
	}
)

func basicAuth(r *http.Request) (username, password string, ok bool) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return
	}
	return parseAuth(auth)
}

func parseAuth(auth string) (username, password string, ok bool) {
	if !strings.HasPrefix(auth, "Basic ") {
		return
	}

	c, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

func requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
	http.Error(w, "Not Authorized", http.StatusUnauthorized)
}

func checkAuth(user, pass string) bool {
	for u, p := range admitUsers {
		if u == user && p == pass {
			return true
		}
	}
	return false
}

func NewBasicAuth() negroni.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		user, pass, ok := basicAuth(r)
		if !ok {
			requireAuth(rw)
			return
		}

		if !checkAuth(user, pass) {
			requireAuth(rw)
			return
		}

		req := rw.(negroni.ResponseWriter)
		if req.Status() != http.StatusUnauthorized {
			next(rw, r)
		}
	}
}
