package userreassign

import (
	"context"

	"github.com/user/reviewer-svc/internal/domain"
	prdomain "github.com/user/reviewer-svc/internal/domain/pr"
	domainuser "github.com/user/reviewer-svc/internal/domain/user"
)

type ReassignmentPRRepository interface {
	ListAssignedTo(ctx context.Context, tx domain.Tx, userID string, status *prdomain.PRStatus) ([]prdomain.PullRequest, error)
	ReplaceReviewers(ctx context.Context, tx domain.Tx, prID string, reviewers []prdomain.PRReviewer) error
}

type ReassignmentUserRepository interface {
	ListActiveByTeamExcept(ctx context.Context, tx domain.Tx, teamID string, exclude []string) ([]domainuser.User, error)
}

type userReassignmentService struct {
	prs   ReassignmentPRRepository
	users ReassignmentUserRepository
	clk   domain.Clock
	strat domainuser.AssignmentStrategy
}

func NewUserReassignmentService(prs ReassignmentPRRepository, users ReassignmentUserRepository, clk domain.Clock, strat domainuser.AssignmentStrategy) domainuser.UserReassignmentService {
	return &userReassignmentService{
		prs:   prs,
		users: users,
		clk:   clk,
		strat: strat,
	}
}

func (s *userReassignmentService) ReassignUserInOpenPRs(ctx context.Context, tx domain.Tx, teamID string, u *domainuser.User) (int, error) {
	open := prdomain.PRStatusOpen

	prs, err := s.prs.ListAssignedTo(ctx, tx, u.ID, &open)
	if err != nil {
		return 0, err
	}

	baseCandidates, err := s.users.ListActiveByTeamExcept(ctx, tx, teamID, []string{u.ID})
	if err != nil {
		return 0, err
	}

	reassigned := 0

	for _, pr := range prs {
		exclude := pr.BuildExcludeList(u.ID)

		cands := make([]domainuser.User, 0, len(baseCandidates))
		for _, cand := range baseCandidates {
			skip := false
			for _, ex := range exclude {
				if cand.ID == ex {
					skip = true
					break
				}
			}
			if !skip {
				cands = append(cands, cand)
			}
		}

		cand, err := s.strat.ChooseReassignment(ctx, *u, cands)
		if err != nil {
			return 0, err
		}

		now := s.clk.Now()
		newReviewers, _ := pr.ReplaceReviewer(u.ID, cand.ID, now)
		prdomain.NormalizeReviewerSlots(newReviewers)

		if err := s.prs.ReplaceReviewers(ctx, tx, pr.ID, newReviewers); err != nil {
			return 0, err
		}
		reassigned++
	}

	return reassigned, nil
}
