package stats

import (
	"context"

	domainstats "github.com/user/reviewer-svc/internal/domain/stats"
)

type Service interface {
	StatsByUser(ctx context.Context, teamID *string) ([]domainstats.UserAssignmentsStats, error)
	StatsByPR(ctx context.Context, teamID *string) ([]domainstats.PRAssignmentsStats, error)
}
