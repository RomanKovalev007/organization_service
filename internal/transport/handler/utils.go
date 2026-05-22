package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/RomanKovalev007/organization_service/internal/apperr"
)

var codeToStatus = map[string]int{
	apperr.CodeInternalError: http.StatusInternalServerError,
	apperr.CodeInvalidInput:  http.StatusBadRequest,
	apperr.CodeAlreadyExists: http.StatusConflict,
	apperr.CodeNotFound:      http.StatusNotFound,
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, apperr.Error{Code: code, Message: msg})
}

func handleAppErr(w http.ResponseWriter, err error) {
	var svcErr *apperr.Error
	if errors.As(err, &svcErr) {
		status, ok := codeToStatus[svcErr.Code]
		if !ok {
			status = http.StatusInternalServerError
		}
		writeError(w, status, svcErr.Code, svcErr.Message)
		return
	}
	writeError(w, http.StatusInternalServerError, apperr.CodeInternalError, "internal server error")
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}
