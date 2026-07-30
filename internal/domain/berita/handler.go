package berita

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/middleware"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/response"
	stor "github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/storage"
)

type Handler struct {
	svc    *Service
	store  stor.Client
}

func NewHandler(svc *Service, store stor.Client) *Handler {
	return &Handler{svc: svc, store: store}
}

func (h *Handler) Register(mux *http.ServeMux, authMw func(http.Handler) http.Handler, roleMw func(http.Handler) http.Handler) {
	mux.Handle("POST /api/v1/berita", authMw(roleMw(http.HandlerFunc(h.Create))))
	mux.Handle("GET /api/v1/berita", authMw(roleMw(http.HandlerFunc(h.List))))
	mux.Handle("GET /api/v1/berita/{id}", authMw(roleMw(http.HandlerFunc(h.Get))))
	mux.Handle("PUT /api/v1/berita/{id}", authMw(roleMw(http.HandlerFunc(h.Update))))
	mux.Handle("DELETE /api/v1/berita/{id}", authMw(roleMw(http.HandlerFunc(h.Delete))))
	mux.Handle("POST /api/v1/berita/{id}/image", authMw(roleMw(http.HandlerFunc(h.UploadImage))))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Title == "" || input.Content == "" {
		response.Error(w, http.StatusBadRequest, "title and content are required")
		return
	}

	b, err := h.svc.Create(r.Context(), userID, input.Title, input.Content)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.signImageURL(r.Context(), b)
	response.JSON(w, http.StatusCreated, b)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i := range list {
		h.signImageURL(r.Context(), &list[i])
	}
	response.JSON(w, http.StatusOK, list)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	b, err := h.svc.ByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if b == nil {
		response.Error(w, http.StatusNotFound, "berita not found")
		return
	}
	h.signImageURL(r.Context(), b)
	response.JSON(w, http.StatusOK, b)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	id := r.PathValue("id")

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, err := h.svc.Update(r.Context(), id, userID, input.Title, input.Content)
	if err != nil {
		switch err.Error() {
		case "berita not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "forbidden: not the author":
			response.Error(w, http.StatusForbidden, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	h.signImageURL(r.Context(), b)
	response.JSON(w, http.StatusOK, b)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	id := r.PathValue("id")

	b, err := h.svc.ByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if b == nil {
		response.Error(w, http.StatusNotFound, "berita not found")
		return
	}

	if b.ImageURL != "" {
		path := b.ImageURL
		if strings.Contains(path, "://") {
			if p, err := extractObjectPath(path); err == nil {
				path = p
			}
		}
		_ = h.store.Delete(r.Context(), path)
	}

	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		switch err.Error() {
		case "berita not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "forbidden: not the author":
			response.Error(w, http.StatusForbidden, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	beritaID := r.PathValue("id")

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "file too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "missing image field")
		return
	}
	defer file.Close()

	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to read image")
		return
	}
	file.Seek(0, io.SeekStart)

	detected := http.DetectContentType(buf)
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowed[detected] {
		response.Error(w, http.StatusBadRequest, "unsupported image type: jpeg, png, gif, webp allowed")
		return
	}

	ext := filepath.Ext(header.Filename)
	validExts := map[string]string{
		".jpg":  ".jpg",
		".jpeg": ".jpeg",
		".png":  ".png",
		".gif":  ".gif",
		".webp": ".webp",
	}
	if _, ok := validExts[strings.ToLower(ext)]; !ok {
		ext = ".jpg"
	}

	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate unique id")
		return
	}
	objectPath := fmt.Sprintf("berita/%s/%s%s", beritaID, hex.EncodeToString(uuid), strings.ToLower(ext))

	imageURL, err := h.store.Upload(r.Context(), objectPath, detected, file)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to upload image")
		return
	}

	b, err := h.svc.SetImage(r.Context(), beritaID, userID, imageURL)
	if err != nil {
		switch err.Error() {
		case "berita not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "forbidden: not the author":
			response.Error(w, http.StatusForbidden, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.signImageURL(r.Context(), b)
	response.JSON(w, http.StatusOK, map[string]string{"image_url": b.ImageURL})
}

func (h *Handler) signImageURL(ctx context.Context, b *Berita) {
	if b.ImageURL == "" {
		return
	}
	signed, err := h.store.PresignGet(ctx, b.ImageURL, 24*time.Hour)
	if err != nil {
		b.ImageURL = ""
		return
	}
	b.ImageURL = signed
}

func extractObjectPath(imageURL string) (string, error) {
	parsed, err := url.Parse(imageURL)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(parsed.Path, "/"), nil
}