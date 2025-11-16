package postgres

import (
	"context"

	domain "github.com/user/reviewer-svc/internal/domain"
	domainteam "github.com/user/reviewer-svc/internal/domain/team"
)

type TeamRepo struct{}

func NewTeamRepo() *TeamRepo {
	return &TeamRepo{}
}

func (r *TeamRepo) Create(ctx context.Context, ttx domain.Tx, t *domainteam.Team) error {
	_, err := ttx.Exec(ctx,
		"INSERT INTO teams (id, name, created_at) VALUES ($1, $2, $3)",
		t.ID, t.Name, t.CreatedAt,
	)
	return translateError(err)
}

func (r *TeamRepo) GetByID(ctx context.Context, ttx domain.Tx, id string) (*domainteam.Team, error) {
	row := ttx.QueryRow(ctx,
		"SELECT id, name, created_at FROM teams WHERE id = $1",
		id,
	)
	var t domainteam.Team
	if err := row.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
		return nil, translateError(err)
	}
	return &t, nil
}

func (r *TeamRepo) List(ctx context.Context, ttx domain.Tx) ([]domainteam.Team, error) {
	rows, err := ttx.Query(ctx,
		"SELECT id, name, created_at FROM teams ORDER BY created_at",
	)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	var res []domainteam.Team
	for rows.Next() {
		var t domainteam.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

var _ domainteam.Repository = (*TeamRepo)(nil)
