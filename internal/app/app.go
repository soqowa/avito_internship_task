package app

import (
	"net/http"

	"log/slog"

	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	handler "github.com/user/reviewer-svc/internal/app/handler"
	prsvc "github.com/user/reviewer-svc/internal/domain/pr"
	statssvc "github.com/user/reviewer-svc/internal/domain/stats"
	teamsvc "github.com/user/reviewer-svc/internal/domain/team"
	usersvc "github.com/user/reviewer-svc/internal/domain/user"
	userreassign "github.com/user/reviewer-svc/internal/domain/userreassign"
	"github.com/user/reviewer-svc/internal/infrastructure/clock"
	postgres "github.com/user/reviewer-svc/internal/infrastructure/db/postgres"
	"github.com/user/reviewer-svc/internal/infrastructure/idgen"
	"github.com/user/reviewer-svc/internal/infrastructure/random"
)

func NewHandler(r chi.Router, pool *pgxpool.Pool, log *slog.Logger) http.Handler {
	txManager := postgres.NewTxManager(pool)
	teamRepo := postgres.NewTeamRepo()
	userRepo := postgres.NewUserRepo()
	prRepo := postgres.NewPRRepo()

	clk := clock.SystemClock{}
	rnd := random.New()
	idGen := idgen.NewUUIDGenerator()
	strategy := usersvc.NewRandomAssignmentStrategy(rnd)

	teamSvc := teamsvc.NewTeamService(teamRepo, txManager, clk, idGen)

	userReassignSvc := userreassign.NewUserReassignmentService(prRepo, userRepo, clk, strategy)
	userSvc := usersvc.NewUserService(userRepo, teamRepo, txManager, clk, idGen, userReassignSvc)
	userBulkSvc := usersvc.NewUserBulkService(userRepo, teamRepo, txManager, userReassignSvc)
	prSvc := prsvc.NewPRService(prRepo, userRepo, txManager, clk, idGen, strategy)
	statsSvc := statssvc.NewStatsService(prRepo, txManager)

	deps := handler.Deps{
		Teams:    teamSvc,
		Users:    userSvc,
		UserBulk: userBulkSvc,
		PRs:      prSvc,
		Stats:    statsSvc,
		Log:      log,
		DB:       pool,
	}

	return handler.NewRouter(r, deps)
}
