package handler

import (
	"context"
	"net/http"

	"log/slog"

	chi "github.com/go-chi/chi/v5"

	"github.com/user/reviewer-svc/internal/app/handler/health"
	"github.com/user/reviewer-svc/internal/app/handler/prs"
	"github.com/user/reviewer-svc/internal/app/handler/stats"
	"github.com/user/reviewer-svc/internal/app/handler/teams"
	"github.com/user/reviewer-svc/internal/app/handler/users"
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

type Deps struct {
	Teams    teams.Service
	Users    users.Service
	UserBulk users.BulkService
	PRs      prs.Service
	Stats    stats.Service
	Log      *slog.Logger
	DB       DBPinger
}

func NewRouter(r chi.Router, d Deps) http.Handler {
	healthHandler := health.NewHandler(d.Log, d.DB)
	teamHandler := teams.NewHandler(d.Teams, d.Log)
	userHandler := users.NewHandler(d.Users, d.UserBulk, d.Log)
	prHandler := prs.NewHandler(d.PRs, d.Log)
	statsHandler := stats.NewHandler(d.Stats, d.Log)

	r.Get("/healthz", healthHandler.Healthz)
	r.Get("/readyz", healthHandler.Readyz)

	r.Route("/teams", func(r chi.Router) {
		r.Post("/", teamHandler.CreateTeam)
		r.Get("/", teamHandler.ListTeams)
		r.Get("/{teamId}", teamHandler.GetTeam)
		r.Post("/{teamId}/users", userHandler.CreateUser)
		r.Post("/{teamId}/deactivate-users", userHandler.BulkDeactivateUsers)
	})

	r.Route("/users", func(r chi.Router) {
		r.Get("/", userHandler.ListUsers)
		r.Get("/{userId}", userHandler.GetUser)
		r.Patch("/{userId}", userHandler.UpdateUser)
		r.Get("/{userId}/assigned-prs", prHandler.ListAssignedPRs)
	})

	r.Route("/prs", func(r chi.Router) {
		r.Post("/", prHandler.CreatePR)
		r.Get("/", prHandler.ListPRs)
		r.Get("/{prId}", prHandler.GetPR)
		r.Post("/{prId}/merge", prHandler.MergePR)
		r.Post("/{prId}/reassign", prHandler.ReassignReviewer)
	})

	r.Route("/stats", func(r chi.Router) {
		r.Get("/assignments", statsHandler.GetAssignmentsStats)
	})

	return r
}
