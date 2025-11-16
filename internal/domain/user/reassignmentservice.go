package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/user/reviewer-svc/internal/domain"
)



type UserReassignmentService interface {
	ReassignUserInOpenPRs(ctx context.Context, tx domain.Tx, teamID uuid.UUID, u *User) (int, error)
}
