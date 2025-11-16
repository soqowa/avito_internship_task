package prs

import (
	"context"

	"github.com/google/uuid"

	domainpr "github.com/user/reviewer-svc/internal/domain/pr"
)

type Service interface {
	CreatePR(ctx context.Context, title string, authorID uuid.UUID) (*domainpr.PullRequest, error)
	GetPR(ctx context.Context, id uuid.UUID) (*domainpr.PullRequest, error)
	ListPRs(ctx context.Context, status *domainpr.PRStatus) ([]domainpr.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID uuid.UUID) (*domainpr.PullRequest, error)
	MergePR(ctx context.Context, prID uuid.UUID) (*domainpr.PullRequest, error)
	ListAssignedPRs(ctx context.Context, userID uuid.UUID, status *domainpr.PRStatus) ([]domainpr.PullRequest, error)
}
