package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/user/reviewer-svc/internal/domain"
	"github.com/user/reviewer-svc/internal/domain/team"
)

type UserRepository interface {
	Create(ctx context.Context, tx domain.Tx, u *User) error
	GetByID(ctx context.Context, tx domain.Tx, id uuid.UUID) (*User, error)
	Update(ctx context.Context, tx domain.Tx, u *User) error
	List(ctx context.Context, tx domain.Tx, teamID *uuid.UUID, isActive *bool) ([]User, error)
}

type TeamRepository interface {
	GetByID(ctx context.Context, tx domain.Tx, id uuid.UUID) (*team.Team, error)
}

type UserService struct {
	users        UserRepository
	teams        TeamRepository
	tx           domain.TxManager
	clk          domain.Clock
	reassignment UserReassignmentService
}

func NewUserService(users UserRepository, teams TeamRepository, tx domain.TxManager, clk domain.Clock, reassignment UserReassignmentService) *UserService {
	return &UserService{
		users:        users,
		teams:        teams,
		tx:           tx,
		clk:          clk,
		reassignment: reassignment,
	}
}

func (s UserService) CreateUser(ctx context.Context, teamID uuid.UUID, name string, isActive bool) (*User, error) {
	if name == "" {
		return nil, domain.ErrInvalidUserName
	}

	user := &User{
		ID:        uuid.New(),
		Name:      name,
		TeamID:    teamID,
		IsActive:  isActive,
		CreatedAt: s.clk.Now(),
	}

	var res *User
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		if _, err := s.teams.GetByID(ctx, ttx, teamID); err != nil {
			return err
		}
		if err := s.users.Create(ctx, ttx, user); err != nil {
			return err
		}
		res = user
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s UserService) ListUsers(ctx context.Context, teamID *uuid.UUID, isActive *bool) ([]User, error) {
	var res []User
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		list, err := s.users.List(ctx, ttx, teamID, isActive)
		if err != nil {
			return err
		}
		res = list
		return nil
	})
	return res, err
}

func (s UserService) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	var res *User
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		u, err := s.users.GetByID(ctx, ttx, id)
		if err != nil {
			return err
		}
		res = u
		return nil
	})
	return res, err
}

func (s UserService) UpdateUser(ctx context.Context, id uuid.UUID, name *string, isActive *bool) (*User, error) {
	if name == nil && isActive == nil {
		return nil, domain.ErrEmptyUpdate
	}
	if name != nil && *name == "" {
		return nil, domain.ErrInvalidUserName
	}

	var res *User
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		u, err := s.users.GetByID(ctx, ttx, id)
		if err != nil {
			return err
		}


		newName := u.Name
		if name != nil {
			newName = *name
		}
		newIsActive := u.IsActive
		if isActive != nil {
			newIsActive = *isActive
		}



		if u.IsActive && !newIsActive {
			if _, err := s.reassignment.ReassignUserInOpenPRs(ctx, ttx, u.TeamID, u); err != nil {
				return err
			}
		}

		u.Name = newName
		u.IsActive = newIsActive

		if err := s.users.Update(ctx, ttx, u); err != nil {
			return err
		}
		res = u
		return nil
	})
	return res, err
}
