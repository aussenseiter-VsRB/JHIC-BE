package middleware

import (
	"net/http"
)

const (
	allowOriginHeader   = "Access-Control-Allow-Origin"
	allowMethodsHeader  = "Access-Control-Allow-Methods"
	allowHeadersHeader  = "Access-Control-Allow-Headers"
	allowCredentials    = "Access-Control-Allow-Credentials"
	exposeHeadersHeader = "Access-Control-Expose-Headers"
	varyHeader          = "Vary"
)

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := false
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
		}
		allowed[o] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowedOrigin := ""
			switch {
			case allowAll && origin == "":
				allowedOrigin = "*"
			case allowAll:
				allowedOrigin = origin
			case origin != "":
				if _, ok := allowed[origin]; ok {
					allowedOrigin = origin
				}
			}

			if allowedOrigin != "" {
				w.Header().Set(allowOriginHeader, allowedOrigin)
				w.Header().Set(allowMethodsHeader, "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set(allowHeadersHeader, "Content-Type, Authorization")
				w.Header().Set(exposeHeadersHeader, "Location")
				if allowedOrigin != "*" {
					w.Header().Set(allowCredentials, "true")
					w.Header().Add(varyHeader, "Origin")
				}
			}

			if r.Method == http.MethodOptions {
				if allowedOrigin == "" {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
