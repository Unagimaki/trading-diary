package httptransport

import "net/http"

type errorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{
		Code:      errorCode(status, message),
		Message:   message,
		RequestID: w.Header().Get("X-Request-ID"),
	})
}

func errorCode(status int, message string) string {
	switch message {
	case "image is too large":
		return "IMAGE_TOO_LARGE"
	case "image is required", "invalid image":
		return "INVALID_IMAGE"
	case "unsupported image type":
		return "IMAGE_UNSUPPORTED_TYPE"
	case "image cell not found", "image not found":
		return "IMAGE_NOT_FOUND"
	case "authentication required":
		return "AUTHENTICATION_REQUIRED"
	case "invalid credentials":
		return "INVALID_CREDENTIALS"
	case "email already registered":
		return "EMAIL_ALREADY_REGISTERED"
	case "invalid request":
		return "INVALID_REQUEST"
	case "invalid journal name":
		return "INVALID_JOURNAL_NAME"
	case "invalid journal settings":
		return "INVALID_JOURNAL_SETTINGS"
	case "journal not found":
		return "JOURNAL_NOT_FOUND"
	case "invalid column":
		return "INVALID_COLUMN"
	case "column not found":
		return "COLUMN_NOT_FOUND"
	case "invalid cell value":
		return "INVALID_CELL_VALUE"
	case "row or column not found":
		return "CELL_NOT_FOUND"
	case "internal error":
		return "INTERNAL_ERROR"
	}
	if status == http.StatusNotFound {
		return "NOT_FOUND"
	}
	if status == http.StatusBadRequest {
		return "INVALID_REQUEST"
	}
	return "REQUEST_FAILED"
}
