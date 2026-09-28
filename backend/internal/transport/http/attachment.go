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

func (h attachmentHandler) upload(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	var err error
	r.Body = http.MaxBytesReader(w, r.Body, attachment.MaxSize+(1<<20))
	if err = r.ParseMultipartForm(attachment.MaxSize); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(w, http.StatusRequestEntityTooLarge, "image is too large")
		} else {
			writeError(w, http.StatusBadRequest, "invalid request")
		}
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "image is required")
		return
	}
	defer file.Close()
	if header.Size > attachment.MaxSize {
		writeError(w, http.StatusRequestEntityTooLarge, "image is too large")
		return
	}
	v := mux.Vars(r)
	item, err := h.attachments.Upload(r.Context(), u.ID, v["journalID"], v["rowID"], v["columnID"], header.Filename, header.Size, file)
	if errors.Is(err, attachment.ErrInvalidFile) {
		writeError(w, 400, "invalid image")
		return
	}
	if errors.Is(err, attachment.ErrTooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, "image is too large")
		return
	}
	if errors.Is(err, attachment.ErrUnsupportedType) {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported image type")
		return
	}
	if errors.Is(err, attachment.ErrNotFound) {
		writeError(w, 404, "image cell not found")
		return
	}
	if err != nil {
		internalError(w, r, "attachment.upload", err, "user_id", u.ID, "journal_id", v["journalID"], "row_id", v["rowID"], "column_id", v["columnID"])
		return
	}
	writeJSON(w, 201, item)
}
func (h attachmentHandler) content(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	var err error
	item, reader, err := h.attachments.Open(r.Context(), u.ID, mux.Vars(r)["id"])
	if errors.Is(err, attachment.ErrNotFound) {
		writeError(w, 404, "image not found")
		return
	}
	if err != nil {
		internalError(w, r, "attachment.open", err, "user_id", u.ID, "attachment_id", mux.Vars(r)["id"])
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
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	var err error
	v := mux.Vars(r)
	err = h.attachments.Delete(r.Context(), u.ID, v["journalID"], v["rowID"], v["columnID"])
	if errors.Is(err, attachment.ErrNotFound) {
		writeError(w, 404, "image not found")
		return
	}
	if err != nil {
		internalError(w, r, "attachment.delete", err, "user_id", u.ID, "journal_id", v["journalID"], "row_id", v["rowID"], "column_id", v["columnID"])
		return
	}
	w.WriteHeader(204)
}
