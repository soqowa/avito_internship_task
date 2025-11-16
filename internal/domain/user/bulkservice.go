package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/user/reviewer-svc/internal/domain"
	"github.com/user/reviewer-svc/internal/domain/team"
)


type UserBulkService struct {
	users        BulkUserRepository
	teams        BulkTeamRepository
	tx           domain.TxManager
	reassignment UserReassignmentService
}

type BulkUserRepository interface {
	ListByIDs(ctx context.Context, tx domain.Tx, ids []uuid.UUID) ([]User, error)
	Update(ctx context.Context, tx domain.Tx, u *User) error
}

type BulkTeamRepository interface {
	GetByID(ctx context.Context, tx domain.Tx, id uuid.UUID) (*team.Team, error)
}

func NewUserBulkService(users BulkUserRepository, teams BulkTeamRepository, tx domain.TxManager, reassignment UserReassignmentService) *UserBulkService {
	return &UserBulkService{
		users:        users,
		teams:        teams,
		tx:           tx,
		reassignment: reassignment,
	}
}


func (s UserBulkService) BulkDeactivate(ctx context.Context, teamID uuid.UUID, userIDs []uuid.UUID) (int, int, error) {
	if len(userIDs) == 0 {
		return 0, 0, domain.ErrEmptyBulkUserIDs
	}



	uniqueUserIDs := make([]uuid.UUID, 0, len(userIDs))
	seen := make(map[uuid.UUID]struct{}, len(userIDs))
	for _, id := range userIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueUserIDs = append(uniqueUserIDs, id)
	}

	var deactivated, reassigned int

	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		if _, err := s.teams.GetByID(ctx, ttx, teamID); err != nil {
			return err
		}

		users, err := s.users.ListByIDs(ctx, ttx, uniqueUserIDs)
		if err != nil {
			return err
		}
		if len(users) != len(uniqueUserIDs) {
			return domain.ErrNotFound
		}

		usersMap := make(map[uuid.UUID]*User, len(users))
		for i := range users {
			u := users[i]
			if u.TeamID != teamID {
				return domain.ErrCrossTeamDeactive
			}
			user := User(u)
			usersMap[user.ID] = &user
		}

		for _, u := range usersMap {
			if !u.IsActive {
				continue
			}
			u.IsActive = false
			if err := s.users.Update(ctx, ttx, u); err != nil {
				return err
			}
			deactivated++

			reassignedForUser, err := s.reassignment.ReassignUserInOpenPRs(ctx, ttx, teamID, u)
			if err != nil {
				return err
			}
			reassigned += reassignedForUser
		}
		return nil
	})
	if err != nil {
		return 0, 0, err
	}
	return deactivated, reassigned, nil
}
