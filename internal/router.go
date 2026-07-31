package internal

import (
	"net/http"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/berita"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/pkl"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/middleware"
)

func NewRouter(ah *auth.Handler, uh *user.Handler, bh *berita.Handler, pklHnd *pkl.Handler, authMw func(http.Handler) http.Handler, roleMw func(http.Handler) http.Handler, roleCheck middleware.RoleChecker) http.Handler {
	mux := http.NewServeMux()

	ah.Register(mux)
	uh.Register(mux, authMw, roleCheck)
	bh.Register(mux, authMw, roleMw)
	pklHnd.Register(mux, authMw, roleCheck)

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	var h http.Handler = mux
	h = middleware.Logger(h)
	h = middleware.CORS(h)
	return h
}
