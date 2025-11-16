package httpserver

	import (
		"encoding/json"
		"errors"
		"net/http"

		"github.com/user/reviewer-svc/internal/domain"
	)
	
	type ErrorResponse struct {
		Error ErrorBody `json:"error"`
	}
	
	type ErrorBody struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

	func WriteError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
		resp := ErrorResponse{Error: ErrorBody{Code: code, Message: message}}
		WriteJSON(w, status, resp)
	}

func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func MapError(err error) (int, string) {
	if err == nil {
		return http.StatusOK, "OK"
	}
	if errors.Is(err, domain.ErrNotFound) {
		return http.StatusNotFound, "NOT_FOUND"
	}
	if errors.Is(err, domain.ErrAlreadyExists) {
		return http.StatusConflict, "PR_EXISTS"
	}
	if errors.Is(err, domain.ErrAlreadyMerged) {
		return http.StatusConflict, "PR_MERGED"
	}
	if errors.Is(err, domain.ErrNoCandidate) {
		return http.StatusConflict, "NO_CANDIDATE"
	}
	if errors.Is(err, domain.ErrBadReviewer) {
		return http.StatusConflict, "NOT_ASSIGNED"
	}
	if errors.Is(err, domain.ErrInvalidTeamName) {
		return http.StatusBadRequest, "INVALID_TEAM_NAME"
	}
	if errors.Is(err, domain.ErrInvalidUserName) {
		return http.StatusBadRequest, "INVALID_USER_NAME"
	}
	if errors.Is(err, domain.ErrInvalidPRTitle) {
		return http.StatusBadRequest, "INVALID_PR_TITLE"
	}
	if errors.Is(err, domain.ErrEmptyUpdate) {
		return http.StatusBadRequest, "EMPTY_UPDATE"
	}
	if errors.Is(err, domain.ErrEmptyBulkUserIDs) {
		return http.StatusBadRequest, "EMPTY_BULK_USER_IDS"
	}
	if errors.Is(err, domain.ErrCrossTeamDeactive) {
		return http.StatusBadRequest, "CROSS_TEAM_DEACTIVATION"
	}
	if errors.Is(err, domain.ErrConstraintViolation) {
		return http.StatusBadRequest, "CONSTRAINT_VIOLATION"
	}
	if errors.Is(err, domain.ErrInvalidRequest) {
		return http.StatusBadRequest, "BAD_REQUEST"
	}
	return http.StatusInternalServerError, "INTERNAL_ERROR"
}
