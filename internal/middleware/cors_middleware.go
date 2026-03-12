package middleware

import "net/http"

var allowedOrigins = map[string]struct{}{
	"http://localhost:5173": {},
	"http://127.0.0.1:5173": {},
}

const (
	allowedMethods = "GET, POST, PATCH, PUT, DELETE, OPTIONS"
	allowedHeaders = "Content-Type, Authorization"
)

func CORSMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			_, isAllowedOrigin := allowedOrigins[origin]

			if isAllowedOrigin {
				headers := w.Header()
				headers.Set("Access-Control-Allow-Origin", origin)
				headers.Add("Vary", "Origin")
				headers.Add("Vary", "Access-Control-Request-Method")
				headers.Add("Vary", "Access-Control-Request-Headers")
				headers.Set("Access-Control-Allow-Methods", allowedMethods)
				headers.Set("Access-Control-Allow-Headers", allowedHeaders)
				headers.Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				if isAllowedOrigin {
					w.WriteHeader(http.StatusNoContent)
				} else {
					http.Error(w, "origin not allowed", http.StatusForbidden)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
