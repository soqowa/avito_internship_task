package prs

import domainpr "github.com/user/reviewer-svc/internal/domain/pr"

func toResponse(pr domainpr.PullRequest) PullRequest {
	res := PullRequest{
		ID:        pr.ID,
		Title:     pr.Title,
		AuthorID:  pr.AuthorID,
		Status:    PRStatus(pr.Status),
		CreatedAt: pr.CreatedAt,
		MergedAt:  pr.MergedAt,
		Reviewers: make([]PRReviewer, 0, len(pr.Reviewers)),
	}
	for _, rv := range pr.Reviewers {
		res.Reviewers = append(res.Reviewers, PRReviewer{
			UserID:     rv.UserID,
			Slot:       rv.Slot,
			AssignedAt: rv.AssignedAt,
		})
	}
	return res
}
