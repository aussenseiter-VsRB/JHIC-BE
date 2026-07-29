package internal

import (
	"net/http"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/pipeline"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/user"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/workspace"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/middleware"
)

func NewRouter(uh *user.Handler, wh *workspace.Handler, ph *pipeline.Handler) http.Handler {
	mux := http.NewServeMux()

	uh.Register(mux)
	wh.Register(mux)
	ph.Register(mux)

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	var h http.Handler = mux
	h = middleware.Logger(h)
	h = middleware.CORS(h)
	return h
}
