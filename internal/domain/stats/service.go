package stats

import (
	"context"

	"github.com/user/reviewer-svc/internal/domain"
)

type PullRequestStatsRepository interface {
	StatsByUser(ctx context.Context, tx domain.Tx, teamID *string) ([]UserAssignmentsStats, error)
	StatsByPR(ctx context.Context, tx domain.Tx, teamID *string) ([]PRAssignmentsStats, error)
}

type StatsService struct {
	prs PullRequestStatsRepository
	tx  domain.TxManager
}

func NewStatsService(prs PullRequestStatsRepository, tx domain.TxManager) *StatsService {
	return &StatsService{prs: prs, tx: tx}
}

func (s StatsService) StatsByUser(ctx context.Context, teamID *string) ([]UserAssignmentsStats, error) {
	var res []UserAssignmentsStats
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		stats, err := s.prs.StatsByUser(ctx, ttx, teamID)
		if err != nil {
			return err
		}
		res = stats
		return nil
	})
	return res, err
}

func (s StatsService) StatsByPR(ctx context.Context, teamID *string) ([]PRAssignmentsStats, error) {
	var res []PRAssignmentsStats
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		stats, err := s.prs.StatsByPR(ctx, ttx, teamID)
		if err != nil {
			return err
		}
		res = stats
		return nil
	})
	return res, err
}
