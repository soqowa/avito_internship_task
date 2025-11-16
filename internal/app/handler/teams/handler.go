package teams

import (
	"errors"
	"net/http"

	"log/slog"

	"github.com/user/reviewer-svc/internal/app/httpserver"
	"github.com/user/reviewer-svc/internal/app/handler/users"
	"github.com/user/reviewer-svc/internal/domain"
)

type Handler struct {
	service Service
	users   users.Service
	log     *slog.Logger
}

func NewHandler(service Service, usersSvc users.Service, log *slog.Logger) *Handler {
	return &Handler{service: service, users: usersSvc, log: log}
}

// @Summary     Create team
// @Tags        teams
// @Accept      json
// @Produce     json
// @Param       body    body      CreateTeamRequest   true  "Team payload"
// @Success     201     {object}  CreateTeamResponse
// @Failure     400     {object}  httpserver.ErrorResponse
// @Router      /team/add [post]
func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
		var req CreateTeamRequest
		if err := httpserver.DecodeJSON(r, &req); err != nil {
			h.log.Error("create team: invalid JSON", "err", err)
			httpserver.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
			return
		}
	
		team, err := h.service.CreateTeam(r.Context(), req.TeamName)
		if err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				h.log.Error("create team failed", "err", err, "code", "TEAM_EXISTS")
				httpserver.WriteError(w, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists", nil)
				return
			}
			status, code := httpserver.MapError(err)
			h.log.Error("create team failed", "err", err, "code", code)
			httpserver.WriteError(w, status, code, err.Error(), nil)
			return
		}

		for _, m := range req.Members {
			_, err := h.users.UpsertUserByID(r.Context(), m.UserID, team.ID, m.Username, m.IsActive)
			if err != nil {
				status, code := httpserver.MapError(err)
				h.log.Error("create team: upsert user failed", "err", err, "code", code)
				httpserver.WriteError(w, status, code, err.Error(), nil)
				return
			}
		}

		usersList, err := h.users.ListUsers(r.Context(), &team.ID, nil)
		if err != nil {
			status, code := httpserver.MapError(err)
			h.log.Error("create team: list users failed", "err", err, "code", code)
			httpserver.WriteError(w, status, code, err.Error(), nil)
			return
		}
	
	respTeam := withMembers(toResponse(*team), usersList)
	httpserver.WriteJSON(w, http.StatusCreated, CreateTeamResponse{Team: respTeam})
}

// @Summary     List teams
// @Tags        teams
// @Produce     json
// @Success     200     {array}   Team
// @Router      /teams [get]
func (h *Handler) ListTeams(w http.ResponseWriter, r *http.Request) {
		teams, err := h.service.ListTeams(r.Context())
		if err != nil {
			status, code := httpserver.MapError(err)
		h.log.Error("list teams failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	res := make([]Team, 0, len(teams))
	for _, t := range teams {
		res = append(res, toResponse(t))
	}
	httpserver.WriteJSON(w, http.StatusOK, res)
}

// @Summary     Get team by name
// @Tags        teams
// @Produce     json
// @Param       team_name  query     string  true  "Team name"
// @Success     200        {object}  Team
// @Failure     400        {object}  httpserver.ErrorResponse
// @Failure     404        {object}  httpserver.ErrorResponse
// @Router      /team/get [get]
func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
		teamName := r.URL.Query().Get("team_name")
		if teamName == "" {
			httpserver.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "team_name is required", nil)
			return
		}

		teams, err := h.service.ListTeams(r.Context())
		if err != nil {
			status, code := httpserver.MapError(err)
			h.log.Error("get team failed", "err", err, "code", code)
			httpserver.WriteError(w, status, code, err.Error(), nil)
			return
		}
	
		var found *Team
		for _, t := range teams {
			if t.Name == teamName {
				base := toResponse(t)
				usersList, uerr := h.users.ListUsers(r.Context(), &t.ID, nil)
				if uerr != nil {
					status, code := httpserver.MapError(uerr)
					h.log.Error("get team: list users failed", "err", uerr, "code", code)
					httpserver.WriteError(w, status, code, uerr.Error(), nil)
					return
				}
				resp := withMembers(base, usersList)
				found = &resp
				break
			}
		}

		if found == nil {
			httpserver.WriteError(w, http.StatusNotFound, "NOT_FOUND", "team not found", nil)
			return
		}

		httpserver.WriteJSON(w, http.StatusOK, *found)
	}
