package teams

import "github.com/user/reviewer-svc/internal/domain/team"

func toResponse(t team.Team) Team {
	return Team{
		ID:        t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
	}
}
