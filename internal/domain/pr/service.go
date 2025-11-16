package pr

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/user/reviewer-svc/internal/domain"
	domainuser "github.com/user/reviewer-svc/internal/domain/user"
)

type PullRequestRepository interface {
	Create(ctx context.Context, tx domain.Tx, pr *PullRequest) error
	GetByID(ctx context.Context, tx domain.Tx, id uuid.UUID, forUpdate bool) (*PullRequest, error)
	UpdateStatus(ctx context.Context, tx domain.Tx, id uuid.UUID, status PRStatus, mergedAt *time.Time) error
	ReplaceReviewers(ctx context.Context, tx domain.Tx, prID uuid.UUID, reviewers []PRReviewer) error
	List(ctx context.Context, tx domain.Tx, status *PRStatus) ([]PullRequest, error)
	ListAssignedTo(ctx context.Context, tx domain.Tx, userID uuid.UUID, status *PRStatus) ([]PullRequest, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, tx domain.Tx, id uuid.UUID) (*domainuser.User, error)
	ListActiveByTeamExcept(ctx context.Context, tx domain.Tx, teamID uuid.UUID, exclude []uuid.UUID) ([]domainuser.User, error)
}

type AssignmentStrategy interface {
	ChooseInitialReviewers(ctx context.Context, candidates []domainuser.User, max int) ([]domainuser.User, error)
	ChooseReassignment(ctx context.Context, oldReviewer domainuser.User, candidates []domainuser.User) (domainuser.User, error)
}

type PRService struct {
	prs   PullRequestRepository
	users UserRepository
	tx    domain.TxManager
	clk   domain.Clock
	strat AssignmentStrategy
}

func NewPRService(prs PullRequestRepository, users UserRepository, tx domain.TxManager, clk domain.Clock, strat AssignmentStrategy) *PRService {
	return &PRService{prs: prs, users: users, tx: tx, clk: clk, strat: strat}
}

func (s PRService) CreatePR(ctx context.Context, title string, authorID uuid.UUID) (*PullRequest, error) {
	if title == "" {
		return nil, domain.ErrInvalidPRTitle
	}

	pr := &PullRequest{
		ID:        uuid.New(),
		Title:     title,
		AuthorID:  authorID,
		Status:    PRStatusOpen,
		CreatedAt: s.clk.Now(),
	}

	var res *PullRequest
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		author, err := s.users.GetByID(ctx, ttx, authorID)
		if err != nil {
			return err
		}

		cands, err := s.users.ListActiveByTeamExcept(ctx, ttx, author.TeamID, []uuid.UUID{author.ID})
		if err != nil {
			return err
		}

		selected, err := s.strat.ChooseInitialReviewers(ctx, cands, 2)
		if err != nil {
			return err
		}

		for i, u := range selected {
			pr.Reviewers = append(pr.Reviewers, PRReviewer{
				PRID:       pr.ID,
				Slot:       i + 1,
				UserID:     u.ID,
				AssignedAt: s.clk.Now(),
			})
		}

		if err := s.prs.Create(ctx, ttx, pr); err != nil {
			return err
		}
		res = pr
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s PRService) GetPR(ctx context.Context, id uuid.UUID) (*PullRequest, error) {
	var res *PullRequest
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		pr, err := s.prs.GetByID(ctx, ttx, id, false)
		if err != nil {
			return err
		}
		res = pr
		return nil
	})
	return res, err
}

func (s PRService) ListPRs(ctx context.Context, status *PRStatus) ([]PullRequest, error) {
	var res []PullRequest
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		list, err := s.prs.List(ctx, ttx, status)
		if err != nil {
			return err
		}
		res = list
		return nil
	})
	return res, err
}

func (s PRService) ReassignReviewer(ctx context.Context, prID, oldReviewerID uuid.UUID) (*PullRequest, error) {
	var res *PullRequest
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		pr, err := s.prs.GetByID(ctx, ttx, prID, true)
		if err != nil {
			return err
		}
		if pr.Status != PRStatusOpen {
			return domain.ErrAlreadyMerged
		}

		oldReviewer, err := s.users.GetByID(ctx, ttx, oldReviewerID)
		if err != nil {
			return err
		}

		exclude := pr.BuildExcludeList(oldReviewerID)

		candidates, err := s.users.ListActiveByTeamExcept(ctx, ttx, oldReviewer.TeamID, exclude)
		if err != nil {
			return err
		}
		cand, err := s.strat.ChooseReassignment(ctx, *oldReviewer, candidates)
		if err != nil {
			return err
		}

		now := s.clk.Now()
		newReviewers, replaced := pr.ReplaceReviewer(oldReviewerID, cand.ID, now)
		if !replaced {
			return domain.ErrBadReviewer
		}

		if err := s.prs.ReplaceReviewers(ctx, ttx, pr.ID, newReviewers); err != nil {
			return err
		}
		pr.Reviewers = newReviewers
		res = pr
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s PRService) MergePR(ctx context.Context, prID uuid.UUID) (*PullRequest, error) {
	var res *PullRequest
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		pr, err := s.prs.GetByID(ctx, ttx, prID, true)
		if err != nil {
			return err
		}
		if pr.Status == PRStatusMerged {
			res = pr
			return nil
		}
		mergedAt := s.clk.Now()
		if err := s.prs.UpdateStatus(ctx, ttx, pr.ID, PRStatusMerged, &mergedAt); err != nil {
			return err
		}
		pr.Status = PRStatusMerged
		pr.MergedAt = &mergedAt
		res = pr
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s PRService) ListAssignedPRs(ctx context.Context, userID uuid.UUID, status *PRStatus) ([]PullRequest, error) {
	var res []PullRequest
	err := s.tx.WithTx(ctx, func(ctx context.Context, ttx domain.Tx) error {
		if _, err := s.users.GetByID(ctx, ttx, userID); err != nil {
			return err
		}
		list, err := s.prs.ListAssignedTo(ctx, ttx, userID, status)
		if err != nil {
			return err
		}
		res = list
		return nil
	})
	return res, err
}
