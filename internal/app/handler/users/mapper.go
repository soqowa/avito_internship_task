package users

import "github.com/user/reviewer-svc/internal/domain/user"

func toResponse(u user.User) User {
	return User{
		ID:        u.ID,
		Name:      u.Name,
		TeamID:    u.TeamID,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
	}
}
