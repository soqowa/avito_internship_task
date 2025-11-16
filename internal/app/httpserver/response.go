package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/user/reviewer-svc/internal/domain"
)

type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
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
	resp := ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
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
		return http.StatusOK, "ok"
	}
	if errors.Is(err, domain.ErrNotFound) {
		return http.StatusNotFound, "not_found"
	}
	if errors.Is(err, domain.ErrAlreadyExists) {
		return http.StatusConflict, "already_exists"
	}
	if errors.Is(err, domain.ErrAlreadyMerged) {
		return http.StatusConflict, "already_merged"
	}
	if errors.Is(err, domain.ErrNoCandidate) {
		return http.StatusConflict, "no_candidate"
	}
	if errors.Is(err, domain.ErrBadReviewer) {
		return http.StatusConflict, "bad_reviewer"
	}
	if errors.Is(err, domain.ErrInvalidTeamName) {
		return http.StatusBadRequest, "invalid_team_name"
	}
	if errors.Is(err, domain.ErrInvalidUserName) {
		return http.StatusBadRequest, "invalid_user_name"
	}
	if errors.Is(err, domain.ErrInvalidPRTitle) {
		return http.StatusBadRequest, "invalid_pr_title"
	}
	if errors.Is(err, domain.ErrEmptyUpdate) {
		return http.StatusBadRequest, "empty_update"
	}
	if errors.Is(err, domain.ErrEmptyBulkUserIDs) {
		return http.StatusBadRequest, "empty_bulk_user_ids"
	}
	if errors.Is(err, domain.ErrCrossTeamDeactive) {
		return http.StatusBadRequest, "cross_team_deactivation"
	}
	if errors.Is(err, domain.ErrConstraintViolation) {
		return http.StatusBadRequest, "constraint_violation"
	}
	if errors.Is(err, domain.ErrInvalidRequest) {
		return http.StatusBadRequest, "bad_request"
	}
	return http.StatusInternalServerError, "internal_error"
}
