package teams

import (
	"context"

	"github.com/google/uuid"

	"github.com/user/reviewer-svc/internal/domain/team"
)

type Service interface {
	CreateTeam(ctx context.Context, name string) (*team.Team, error)
	ListTeams(ctx context.Context) ([]team.Team, error)
	GetTeam(ctx context.Context, id uuid.UUID) (*team.Team, error)
}
