package users

import (
	"net/http"
	"strconv"

	"log/slog"

	chi "github.com/go-chi/chi/v5"

	"github.com/user/reviewer-svc/internal/app/httpserver"
)

type Handler struct {
	users Service
	bulk  BulkService
	teams TeamService
	log   *slog.Logger
}

func NewHandler(users Service, bulk BulkService, teams TeamService, log *slog.Logger) *Handler {
	return &Handler{users: users, bulk: bulk, teams: teams, log: log}
}

// @Summary     Set user active status
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       body    body      SetIsActiveRequest   true  "User active status"
// @Success     200     {object}  SetIsActiveResponse
// @Failure     400     {object}  httpserver.ErrorResponse
// @Failure     404     {object}  httpserver.ErrorResponse
// @Router      /users/setIsActive [post]
func (h *Handler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		h.log.Error("set is_active: invalid JSON", "err", err)
		httpserver.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}

	userID := req.UserID

	isActive := req.IsActive
	user, err := h.users.UpdateUser(r.Context(), userID, nil, &isActive)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("set is_active failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	team, err := h.teams.GetTeam(r.Context(), user.TeamID)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("get team failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, SetIsActiveResponse{User: toResponseWithTeam(*user, *team)})
}


// @Summary     Create user in team
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       teamId  path      string             true  "Team ID"
// @Param       body    body      CreateUserRequest  true  "User payload"
// @Success     201     {object}  User
// @Failure     400     {object}  httpserver.ErrorResponse
// @Failure     404     {object}  httpserver.ErrorResponse
// @Router      /teams/{teamId}/users [post]
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamId")

	var req CreateUserRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		h.log.Error("create user: invalid JSON", "err", err)
		httpserver.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON", nil)
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user, err := h.users.CreateUser(r.Context(), teamID, req.Name, isActive)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("create user failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	httpserver.WriteJSON(w, http.StatusCreated, toResponseSimple(*user))
}


// @Summary     List users
// @Tags        users
// @Produce     json
// @Param       teamId   query     string  false  "Team ID"
// @Param       isActive query     bool    false  "Filter by active flag"
// @Success     200      {array}   User
// @Router      /users [get]
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var teamID *string
	if v := q.Get("teamId"); v != "" {
		teamID = &v
	}

	var isActive *bool
	if v := q.Get("isActive"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			httpserver.WriteError(w, http.StatusBadRequest, "bad_request", "invalid isActive", nil)
			return
		}
		isActive = &b
	}

	users, err := h.users.ListUsers(r.Context(), teamID, isActive)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("list users failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	res := make([]User, 0, len(users))
	for _, u := range users {
		res = append(res, toResponseSimple(u))
	}
	httpserver.WriteJSON(w, http.StatusOK, res)
}


// @Summary     Get user by ID
// @Tags        users
// @Produce     json
// @Param       userId  path      string  true  "User ID"
// @Success     200     {object}  User
// @Failure     400     {object}  httpserver.ErrorResponse
// @Failure     404     {object}  httpserver.ErrorResponse
// @Router      /users/{userId} [get]
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")

	user, err := h.users.GetUser(r.Context(), id)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("get user failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, toResponseSimple(*user))
}


// @Summary     Update user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       userId  path      string              true  "User ID"
// @Param       body    body      UpdateUserRequest   true  "User payload"
// @Success     200     {object}  User
// @Failure     400     {object}  httpserver.ErrorResponse
// @Failure     404     {object}  httpserver.ErrorResponse
// @Failure     409     {object}  httpserver.ErrorResponse
// @Router      /users/{userId} [patch]
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")

	var req UpdateUserRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		h.log.Error("update user: invalid JSON", "err", err)
		httpserver.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON", nil)
		return
	}

	user, err := h.users.UpdateUser(r.Context(), id, req.Name, req.IsActive)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("update user failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, toResponseSimple(*user))
}


// @Summary     Bulk deactivate users in team and safely reassign open PRs
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       teamId  path      string                        true  "Team ID"
// @Param       body    body      BulkDeactivateUsersRequest    true  "Bulk payload"
// @Success     200     {object}  BulkDeactivateUsersResponse
// @Failure     400     {object}  httpserver.ErrorResponse
// @Failure     404     {object}  httpserver.ErrorResponse
// @Failure     409     {object}  httpserver.ErrorResponse
// @Router      /teams/{teamId}/deactivate-users [post]
func (h *Handler) BulkDeactivateUsers(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamId")

	var req BulkDeactivateUsersRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		h.log.Error("bulk deactivate: invalid JSON", "err", err)
		httpserver.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON", nil)
		return
	}

	if len(req.UserIDs) == 0 {
		httpserver.WriteError(w, http.StatusBadRequest, "empty_bulk_user_ids", "empty userIds", nil)
		return
	}

	deactivated, reassigned, err := h.bulk.BulkDeactivate(r.Context(), teamID, req.UserIDs)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("bulk deactivate failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	resp := BulkDeactivateUsersResponse{DeactivatedCount: deactivated, ReassignedSlotsCount: reassigned}
	httpserver.WriteJSON(w, http.StatusOK, resp)
}
