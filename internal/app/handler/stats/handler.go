package stats

import (
	"net/http"

	"log/slog"

	"github.com/user/reviewer-svc/internal/app/httpserver"
)
type Handler struct {
	service Service
	log     *slog.Logger
}

func NewHandler(service Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}


// @Summary     Assignments statistics
// @Tags        stats
// @Produce     json
// @Param       by      query     string  true   "Aggregation mode" Enums(user,pr)
// @Param       teamId  query     string  false  "Filter by team ID"
// @Success     200     {object}  UserAssignmentsStatsResponse
// @Failure     400     {object}  httpserver.ErrorResponse
// @Router      /stats/assignments [get]
func (h *Handler) GetAssignmentsStats(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	by := q.Get("by")
	if by != "user" && by != "pr" {
		httpserver.WriteError(w, http.StatusBadRequest, "bad_request", "invalid by", nil)
		return
	}

	var teamID *string
	if v := q.Get("teamId"); v != "" {
		teamID = &v
	}

	ctx := r.Context()

	switch by {
	case "user":
		stats, err := h.service.StatsByUser(ctx, teamID)
		if err != nil {
			status, code := httpserver.MapError(err)
			h.log.Error("stats by user failed", "err", err, "code", code)
			httpserver.WriteError(w, status, code, err.Error(), nil)
			return
		}
		res := UserAssignmentsStatsResponse{Items: make([]UserAssignmentsStatsItem, 0, len(stats))}
		for _, st := range stats {
			item, ok := toResponse(st).(UserAssignmentsStatsItem)
			if !ok {
				continue
			}
			res.Items = append(res.Items, item)
		}
		httpserver.WriteJSON(w, http.StatusOK, res)
	case "pr":
		stats, err := h.service.StatsByPR(ctx, teamID)
		if err != nil {
			status, code := httpserver.MapError(err)
			h.log.Error("stats by pr failed", "err", err, "code", code)
			httpserver.WriteError(w, status, code, err.Error(), nil)
			return
		}
		res := PRAssignmentsStatsResponse{Items: make([]PRAssignmentsStatsItem, 0, len(stats))}
		for _, st := range stats {
			item, ok := toResponse(st).(PRAssignmentsStatsItem)
			if !ok {
				continue
			}
			res.Items = append(res.Items, item)
		}
		httpserver.WriteJSON(w, http.StatusOK, res)
	}
}
