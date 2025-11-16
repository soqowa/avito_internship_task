package teams

import (
	"github.com/user/reviewer-svc/internal/domain/team"
	domainuser "github.com/user/reviewer-svc/internal/domain/user"
)

func toResponse(t team.Team) Team {
	return Team{
		TeamName: t.Name,
		Members:  nil,
	}
}

func withMembers(base Team, users []domainuser.User) Team {
	members := make([]TeamMember, 0, len(users))
	for _, u := range users {
		members = append(members, TeamMember{
			UserID:   u.ID,
			Username: u.Name,
			IsActive: u.IsActive,
		})
	}
	base.Members = members
	return base
}
