package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/Rengoku1926/gpu-inference/internal/errs"
	"github.com/Rengoku1926/gpu-inference/internal/middleware"
)

const maxBodyBytes = 1 << 20

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError is the one place that maps errors to HTTP status codes. Client
// errors return their message; anything unexpected is logged with the full
// error chain and the client only sees a generic message plus the request ID.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	reqID := middleware.GetRequestID(r.Context())

	var status int
	switch {
	case errors.Is(err, errs.ErrValidation):
		status = http.StatusBadRequest
	case errors.Is(err, errs.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, errs.ErrDuplicate):
		status = http.StatusConflict
	default:
		zerolog.Ctx(r.Context()).Error().Err(err).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("unhandled error")
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error", RequestID: reqID})
		return
	}
	writeJSON(w, status, errorResponse{Error: err.Error(), RequestID: reqID})
}

// decodeJSON reads a size limited JSON body and rejects unknown fields.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: invalid JSON body: %s", errs.ErrValidation, err.Error())
	}
	return nil
}
