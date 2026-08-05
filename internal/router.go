package internal

import (
	"net/http"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/analytics"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/auth"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/berita"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/chat"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/cvreview"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/match"
	nexxaspmb "github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa/spmb"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/pkl"
	spmb "github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/spmb"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/middleware"
)

func NewRouter(ah *auth.Handler, uh *user.Handler, bh *berita.Handler, pklHnd *pkl.Handler, chatHnd *chat.Handler, matchHnd *match.Handler, cvHnd *cvreview.Handler, aiHnd *nexxaspmb.Handler, spmbRegHnd *spmb.Handler, authMw func(http.Handler) http.Handler, roleMw func(http.Handler) http.Handler, roleCheck middleware.RoleChecker, corsOrigins []string, analyticsHandlers ...*analytics.Handler) http.Handler {
	mux := http.NewServeMux()

	ah.Register(mux)
	uh.Register(mux, authMw, roleCheck)
	bh.Register(mux, authMw, roleMw)
	pklHnd.Register(mux, authMw, roleCheck)
	chatHnd.Register(mux)
	matchHnd.Register(mux)
	cvHnd.Register(mux)
	aiHnd.Register(mux)
	spmbRegHnd.Register(mux, authMw, roleCheck)
	if len(analyticsHandlers) > 0 && analyticsHandlers[0] != nil {
		analyticsHandlers[0].Register(mux, func(next http.Handler) http.Handler { return authMw(middleware.RequireRole("admin")(roleCheck)(next)) })
		analyticsHandlers[0].RegisterBerita(mux, func(next http.Handler) http.Handler {
			return authMw(middleware.RequireRole("jurnal", "admin")(roleCheck)(next))
		})
	}

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	var h http.Handler = mux
	h = middleware.Logger(h)
	h = middleware.CORS(corsOrigins)(h)
	return h
}
