package httptransport

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/trade-diary/backend/internal/domain/attachment"
	"github.com/trade-diary/backend/internal/domain/auth"
	"io"
	"mime"
	"net/http"
)

type attachmentHandler struct {
	auth        *auth.Service
	attachments *attachment.Service
}

func (h attachmentHandler) user(r *http.Request) (auth.User, error) {
	token, ok := cookieHash(r)
	if !ok {
		return auth.User{}, auth.ErrInvalidCredentials
	}
	return h.auth.CurrentUser(r.Context(), token)
}
func (h attachmentHandler) upload(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, attachment.MaxSize+(1<<20))
	if err = r.ParseMultipartForm(attachment.MaxSize); err != nil {
		writeError(w, 400, "image is too large")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "image is required")
		return
	}
	defer file.Close()
	v := mux.Vars(r)
	item, err := h.attachments.Upload(r.Context(), u.ID, v["journalID"], v["rowID"], v["columnID"], header.Filename, header.Size, file)
	if errors.Is(err, attachment.ErrInvalidFile) {
		writeError(w, 400, "invalid image")
		return
	}
	if errors.Is(err, attachment.ErrNotFound) {
		writeError(w, 404, "image cell not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	writeJSON(w, 201, item)
}
func (h attachmentHandler) content(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	item, reader, err := h.attachments.Open(r.Context(), u.ID, mux.Vars(r)["id"])
	if errors.Is(err, attachment.ErrNotFound) {
		writeError(w, 404, "image not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	defer reader.Close()
	w.Header().Set("Content-Type", item.MimeType)
	w.Header().Set("Content-Length", fmt.Sprint(item.Size))
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": item.Name}))
	w.Header().Set("Cache-Control", "private, max-age=3600")
	_, _ = io.Copy(w, reader)
}
func (h attachmentHandler) delete(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	v := mux.Vars(r)
	err = h.attachments.Delete(r.Context(), u.ID, v["journalID"], v["rowID"], v["columnID"])
	if errors.Is(err, attachment.ErrNotFound) {
		writeError(w, 404, "image not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	w.WriteHeader(204)
}
