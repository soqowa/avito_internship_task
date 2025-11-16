package prs

import (
	"net/http"

	"log/slog"

	chi "github.com/go-chi/chi/v5"

	"github.com/user/reviewer-svc/internal/app/httpserver"
	domainpr "github.com/user/reviewer-svc/internal/domain/pr"
)


type Handler struct {
	service Service
	log     *slog.Logger
}

func NewHandler(service Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}


// @Summary     Create pull request with automatic reviewer assignment
// @Tags        prs
// @Accept      json
// @Produce     json
// @Param       body  body      CreatePRRequest  true  "PR payload"
// @Success     201   {object}  PullRequest
// @Failure     400   {object}  httpserver.ErrorResponse
// @Failure     404   {object}  httpserver.ErrorResponse
// @Router      /prs [post]
func (h *Handler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		h.log.Error("create pr: invalid JSON", "err", err)
		httpserver.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}

	pr, err := h.service.CreatePRByID(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("create pr failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	httpserver.WriteJSON(w, http.StatusCreated, CreatePRResponse{PR: toResponse(*pr)})
}


// @Summary     List pull requests
// @Tags        prs
// @Produce     json
// @Param       status  query     string  false  "PR status (OPEN|MERGED)"
// @Success     200     {array}   PullRequest
// @Router      /prs [get]
func (h *Handler) ListPRs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var status *domainpr.PRStatus
	if v := q.Get("status"); v != "" {
		st := domainpr.PRStatus(v)
		status = &st
	}

	prs, err := h.service.ListPRs(r.Context(), status)
	if err != nil {
		statusCode, code := httpserver.MapError(err)
		h.log.Error("list prs failed", "err", err, "code", code)
		httpserver.WriteError(w, statusCode, code, err.Error(), nil)
		return
	}

	res := make([]PullRequest, 0, len(prs))
	for _, p := range prs {
		res = append(res, toResponse(p))
	}
	httpserver.WriteJSON(w, http.StatusOK, res)
}


// @Summary     Get pull request by ID
// @Tags        prs
// @Produce     json
// @Param       prId  path      string  true  "PR ID"
// @Success     200   {object}  PullRequest
// @Failure     400   {object}  httpserver.ErrorResponse
// @Failure     404   {object}  httpserver.ErrorResponse
// @Router      /prs/{prId} [get]
func (h *Handler) GetPR(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "prId")

	prResult, err := h.service.GetPRByID(r.Context(), id)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("get pr failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, toResponse(*prResult))
}


// @Summary     Reassign reviewer for PR
// @Tags        prs
// @Accept      json
// @Produce     json
// @Param       prId  path      string                    true  "PR ID"
// @Param       body  body      ReassignReviewerRequest   true  "Reassign payload"
// @Success     200   {object}  PullRequest
// @Failure     400   {object}  httpserver.ErrorResponse
// @Failure     404   {object}  httpserver.ErrorResponse
// @Failure     409   {object}  httpserver.ErrorResponse
// @Router      /prs/{prId}/reassign [post]
func (h *Handler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req ReassignReviewerRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		h.log.Error("reassign reviewer: invalid JSON", "err", err)
		httpserver.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}

	pr, newReviewerID, err := h.service.ReassignReviewerByID(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("reassign reviewer failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, ReassignReviewerResponse{
		PR:         toResponse(*pr),
		ReplacedBy: newReviewerID,
	})
}


// @Summary     Merge PR (idempotent)
// @Tags        prs
// @Produce     json
// @Param       prId  path      string  true  "PR ID"
// @Success     200   {object}  PullRequest
// @Failure     400   {object}  httpserver.ErrorResponse
// @Failure     404   {object}  httpserver.ErrorResponse
// @Router      /prs/{prId}/merge [post]
func (h *Handler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req MergePRRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		h.log.Error("merge pr: invalid JSON", "err", err)
		httpserver.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}

	pr, err := h.service.MergePRByID(r.Context(), req.PullRequestID)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("merge pr failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, MergePRResponse{PR: toResponse(*pr)})
}


// @Summary     List PRs assigned to user as reviewer
// @Tags        prs
// @Produce     json
// @Param       userId  path      string  true  "User ID"
// @Param       status  query     string  false "PR status (OPEN|MERGED)"
// @Success     200     {array}   PullRequest
// @Failure     400     {object}  httpserver.ErrorResponse
// @Failure     404     {object}  httpserver.ErrorResponse
// @Router      /users/{userId}/assigned-prs [get]
func (h *Handler) ListAssignedPRs(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		httpserver.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required", nil)
		return
	}

	prs, err := h.service.ListAssignedPRsByID(r.Context(), userIDStr, nil)
	if err != nil {
		statusCode, code := httpserver.MapError(err)
		h.log.Error("list assigned prs failed", "err", err, "code", code)
		httpserver.WriteError(w, statusCode, code, err.Error(), nil)
		return
	}

	prList := make([]PullRequestShort, 0, len(prs))
	for _, p := range prs {
		prList = append(prList, toShortResponse(p))
	}
	
	httpserver.WriteJSON(w, http.StatusOK, GetReviewResponse{
		UserID:       userIDStr,
		PullRequests: prList,
	})
}
