package users

import (
	"context"

	domainteam "github.com/user/reviewer-svc/internal/domain/team"
	domainuser "github.com/user/reviewer-svc/internal/domain/user"
)

type Service interface {
	CreateUser(ctx context.Context, teamID string, name string, isActive bool) (*domainuser.User, error)
	UpsertUserByID(ctx context.Context, userID string, teamID string, name string, isActive bool) (*domainuser.User, error)
	ListUsers(ctx context.Context, teamID *string, isActive *bool) ([]domainuser.User, error)
	GetUser(ctx context.Context, id string) (*domainuser.User, error)
	UpdateUser(ctx context.Context, id string, name *string, isActive *bool) (*domainuser.User, error)
}

type BulkService interface {
	BulkDeactivate(ctx context.Context, teamID string, userIDs []string) (int, int, error)
}

type TeamService interface {
	GetTeam(ctx context.Context, id string) (*domainteam.Team, error)
}
