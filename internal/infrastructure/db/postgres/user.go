package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	domain "github.com/user/reviewer-svc/internal/domain"
	domainuser "github.com/user/reviewer-svc/internal/domain/user"
	userreassign "github.com/user/reviewer-svc/internal/domain/userreassign"
)


type UserRepo struct{}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) Create(ctx context.Context, ttx domain.Tx, u *domainuser.User) error {
	_, err := ttx.Exec(ctx,
		"INSERT INTO users (id, name, team_id, is_active, created_at) VALUES ($1, $2, $3, $4, $5)",
		u.ID, u.Name, u.TeamID, u.IsActive, u.CreatedAt,
	)
	return translateError(err)
}

func (r *UserRepo) GetByID(ctx context.Context, ttx domain.Tx, id uuid.UUID) (*domainuser.User, error) {
	row := ttx.QueryRow(ctx,
		"SELECT id, name, team_id, is_active, created_at FROM users WHERE id = $1",
		id,
	)
	var u domainuser.User
	if err := row.Scan(&u.ID, &u.Name, &u.TeamID, &u.IsActive, &u.CreatedAt); err != nil {
		return nil, translateError(err)
	}
	return &u, nil
}

func (r *UserRepo) Update(ctx context.Context, ttx domain.Tx, u *domainuser.User) error {
	_, err := ttx.Exec(ctx,
		"UPDATE users SET name = $1, is_active = $2 WHERE id = $3",
		u.Name, u.IsActive, u.ID,
	)
	return translateError(err)
}

func (r *UserRepo) List(ctx context.Context, ttx domain.Tx, teamID *uuid.UUID, isActive *bool) ([]domainuser.User, error) {
	query := "SELECT id, name, team_id, is_active, created_at FROM users"
	var args []any
	var conds []string

	if teamID != nil {
		args = append(args, *teamID)
		conds = append(conds, fmt.Sprintf("team_id = $%d", len(args)))
	}
	if isActive != nil {
		args = append(args, *isActive)
		conds = append(conds, fmt.Sprintf("is_active = $%d", len(args)))
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY created_at"

	rows, err := ttx.Query(ctx, query, args...)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	var res []domainuser.User
	for rows.Next() {
		var u domainuser.User
		if err := rows.Scan(&u.ID, &u.Name, &u.TeamID, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *UserRepo) ListByIDs(ctx context.Context, ttx domain.Tx, ids []uuid.UUID) ([]domainuser.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query, args := buildUUIDInQuery("SELECT id, name, team_id, is_active, created_at FROM users WHERE id IN (", ")", ids)

	rows, err := ttx.Query(ctx, query, args...)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	var res []domainuser.User
	for rows.Next() {
		var u domainuser.User
		if err := rows.Scan(&u.ID, &u.Name, &u.TeamID, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *UserRepo) ListActiveByTeamExcept(ctx context.Context, ttx domain.Tx, teamID uuid.UUID, exclude []uuid.UUID) ([]domainuser.User, error) {
	query := "SELECT id, name, team_id, is_active, created_at FROM users WHERE team_id = $1 AND is_active = TRUE"
	args := []any{teamID}

	if len(exclude) > 0 {
		var placeholders []string
		for i, id := range exclude {
			args = append(args, id)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
		}
		query += " AND id NOT IN (" + strings.Join(placeholders, ",") + ")"
	}
	query += " ORDER BY created_at"

	rows, err := ttx.Query(ctx, query, args...)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	var res []domainuser.User
	for rows.Next() {
		var u domainuser.User
		if err := rows.Scan(&u.ID, &u.Name, &u.TeamID, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

var _ domainuser.UserRepository = (*UserRepo)(nil)
var _ domainuser.BulkUserRepository = (*UserRepo)(nil)
var _ userreassign.ReassignmentUserRepository = (*UserRepo)(nil)
