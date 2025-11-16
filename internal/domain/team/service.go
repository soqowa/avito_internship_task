package team

import (
	"context"

	"github.com/user/reviewer-svc/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, tx domain.Tx, t *Team) error
	GetByID(ctx context.Context, tx domain.Tx, id string) (*Team, error)
	List(ctx context.Context, tx domain.Tx) ([]Team, error)
}

type TeamService struct {
	teams Repository
	tx    domain.TxManager
	clk   domain.Clock
	idGen domain.IDGenerator
}

func NewTeamService(teams Repository, tx domain.TxManager, clk domain.Clock, idGen domain.IDGenerator) *TeamService {
	return &TeamService{teams: teams, tx: tx, clk: clk, idGen: idGen}
}

func (s TeamService) CreateTeam(ctx context.Context, name string) (*Team, error) {
	if name == "" {
		return nil, domain.ErrInvalidTeamName
	}

	team := &Team{
		ID:        s.idGen.Generate(),
		Name:      name,
		CreatedAt: s.clk.Now(),
	}

	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		return s.teams.Create(ctx, ttx, team)
	})
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s TeamService) ListTeams(ctx context.Context) ([]Team, error) {
	var res []Team
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		list, err := s.teams.List(ctx, ttx)
		if err != nil {
			return err
		}
		res = list
		return nil
	})
	return res, err
}

func (s TeamService) GetTeam(ctx context.Context, id string) (*Team, error) {
	var res *Team
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		team, err := s.teams.GetByID(ctx, ttx, id)
		if err != nil {
			return err
		}
		res = team
		return nil
	})
	return res, err
}
