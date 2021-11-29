package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/zutrixpog/CMS/db"
)

func CookieMiddleware(db *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetIpFromReq(r)
			userId, _ := r.Cookie("userId")
			if userId == nil {
				userId := db.CreateUser(ip)
				http.SetCookie(w, &http.Cookie{
					Name:     "userId",
					Value:    userId,
					HttpOnly: true,
				})
			}
			ctx := context.WithValue(r.Context(), "userId", userId)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func GetIpFromReq(r *http.Request) string {
	var result string
	forwarded := r.Header.Get("X-Forwarded-for")
	if forwarded != "" {
		result = forwarded
	}
	result = r.RemoteAddr
	return strings.Split(result, ":")[0]
}
