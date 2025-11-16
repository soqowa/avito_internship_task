package prs

import (
	"context"

	domainpr "github.com/user/reviewer-svc/internal/domain/pr"
)

type Service interface {
	CreatePRByID(ctx context.Context, prID, title, authorID string) (*domainpr.PullRequest, error)
	GetPRByID(ctx context.Context, id string) (*domainpr.PullRequest, error)
	ListPRs(ctx context.Context, status *domainpr.PRStatus) ([]domainpr.PullRequest, error)
	ReassignReviewerByID(ctx context.Context, prID, oldReviewerID string) (*domainpr.PullRequest, string, error)
	MergePRByID(ctx context.Context, prID string) (*domainpr.PullRequest, error)
	ListAssignedPRsByID(ctx context.Context, userID string, status *domainpr.PRStatus) ([]domainpr.PullRequest, error)
}
