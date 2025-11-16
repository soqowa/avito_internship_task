package stats

import (
	"context"

	"github.com/google/uuid"

	domainstats "github.com/user/reviewer-svc/internal/domain/stats"
)

type Service interface {
	StatsByUser(ctx context.Context, teamID *uuid.UUID) ([]domainstats.UserAssignmentsStats, error)
	StatsByPR(ctx context.Context, teamID *uuid.UUID) ([]domainstats.PRAssignmentsStats, error)
}
