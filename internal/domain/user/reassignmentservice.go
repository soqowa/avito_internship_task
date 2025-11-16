package user

import (
	"context"

	"github.com/user/reviewer-svc/internal/domain"
)



type UserReassignmentService interface {
	ReassignUserInOpenPRs(ctx context.Context, tx domain.Tx, teamID string, u *User) (int, error)
}
