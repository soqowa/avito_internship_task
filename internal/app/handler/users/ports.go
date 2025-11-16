package users

import (
	"context"

	"github.com/google/uuid"

	domainuser "github.com/user/reviewer-svc/internal/domain/user"
)

type Service interface {
	CreateUser(ctx context.Context, teamID uuid.UUID, name string, isActive bool) (*domainuser.User, error)
	ListUsers(ctx context.Context, teamID *uuid.UUID, isActive *bool) ([]domainuser.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, name *string, isActive *bool) (*domainuser.User, error)
}

type BulkService interface {
	BulkDeactivate(ctx context.Context, teamID uuid.UUID, userIDs []uuid.UUID) (int, int, error)
}
