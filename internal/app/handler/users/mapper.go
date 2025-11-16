package users

import (
	"github.com/user/reviewer-svc/internal/domain/team"
	"github.com/user/reviewer-svc/internal/domain/user"
)

func toResponse(u user.User, teamName string) User {
	return User{
		UserID:   u.ID,
		Username: u.Name,
		TeamName: teamName,
		IsActive: u.IsActive,
	}
}

func toResponseWithTeam(u user.User, t team.Team) User {
	return toResponse(u, t.Name)
}

func toResponseSimple(u user.User) User {
	return User{
		UserID:   u.ID,
		Username: u.Name,
		TeamName: "",
		IsActive: u.IsActive,
	}
}
