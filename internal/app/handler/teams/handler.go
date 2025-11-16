package teams

import (
	"net/http"

	"log/slog"

	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/user/reviewer-svc/internal/app/httpserver"
)

type Handler struct {
	service Service
	log     *slog.Logger
}

func NewHandler(service Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}


// @Summary     Create team
// @Tags        teams
// @Accept      json
// @Produce     json
// @Param       body  body      CreateTeamRequest  true  "Team payload"
// @Success     201   {object}  Team
// @Failure     400   {object}  httpserver.ErrorResponse
// @Failure     409   {object}  httpserver.ErrorResponse
// @Router      /teams [post]
func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		h.log.Error("create team: invalid JSON", "err", err)
		httpserver.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON", nil)
		return
	}

	team, err := h.service.CreateTeam(r.Context(), req.Name)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("create team failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	httpserver.WriteJSON(w, http.StatusCreated, toResponse(*team))
}


// @Summary     List teams
// @Tags        teams
// @Produce     json
// @Success     200   {array}   Team
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


// @Summary     Get team by ID
// @Tags        teams
// @Produce     json
// @Param       teamId  path      string  true  "Team ID"
// @Success     200     {object}  Team
// @Failure     400     {object}  httpserver.ErrorResponse
// @Failure     404     {object}  httpserver.ErrorResponse
// @Router      /teams/{teamId} [get]
func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "teamId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpserver.WriteError(w, http.StatusBadRequest, "bad_request", "invalid team id", nil)
		return
	}

	team, err := h.service.GetTeam(r.Context(), id)
	if err != nil {
		status, code := httpserver.MapError(err)
		h.log.Error("get team failed", "err", err, "code", code)
		httpserver.WriteError(w, status, code, err.Error(), nil)
		return
	}

	httpserver.WriteJSON(w, http.StatusOK, toResponse(*team))
}
