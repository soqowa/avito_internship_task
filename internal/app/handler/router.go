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
	teamHandler := teams.NewHandler(d.Teams, d.Users, d.Log)
	userHandler := users.NewHandler(d.Users, d.UserBulk, d.Teams, d.Log)
	prHandler := prs.NewHandler(d.PRs, d.Log)
	statsHandler := stats.NewHandler(d.Stats, d.Log)

	r.Get("/healthz", healthHandler.Healthz)
	r.Get("/readyz", healthHandler.Readyz)

	r.Route("/team", func(r chi.Router) {
		r.Post("/add", teamHandler.CreateTeam)
		r.Get("/get", teamHandler.GetTeam)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userHandler.SetIsActive)
		r.Get("/getReview", prHandler.ListAssignedPRs)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", prHandler.CreatePR)
		r.Post("/merge", prHandler.MergePR)
		r.Post("/reassign", prHandler.ReassignReviewer)
	})

	r.Route("/stats", func(r chi.Router) {
		r.Get("/assignments", statsHandler.GetAssignmentsStats)
	})

	return r
}
