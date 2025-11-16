package prs

import domainpr "github.com/user/reviewer-svc/internal/domain/pr"

func toResponse(pr domainpr.PullRequest) PullRequest {
	assignedReviewers := make([]string, 0, len(pr.Reviewers))
	for _, rv := range pr.Reviewers {
		assignedReviewers = append(assignedReviewers, rv.UserID)
	}

	var createdAt *string
	if !pr.CreatedAt.IsZero() {
		s := pr.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		createdAt = &s
	}

	var mergedAt *string
	if pr.MergedAt != nil && !pr.MergedAt.IsZero() {
		s := pr.MergedAt.Format("2006-01-02T15:04:05Z07:00")
		mergedAt = &s
	}

	return PullRequest{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Title,
		AuthorID:          pr.AuthorID,
		Status:            PRStatus(pr.Status),
		AssignedReviewers: assignedReviewers,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}
}

func toShortResponse(pr domainpr.PullRequest) PullRequestShort {
	return PullRequestShort{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Title,
		AuthorID:        pr.AuthorID,
		Status:          PRStatus(pr.Status),
	}
}
