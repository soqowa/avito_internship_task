package pr

import (
	"time"
)

func (pr PullRequest) BuildExcludeList(targetUserID string) []string {
	exclude := []string{targetUserID, pr.AuthorID}
	for _, r := range pr.Reviewers {
		if r.UserID != targetUserID {
			exclude = append(exclude, r.UserID)
		}
	}
	return exclude
}

func (pr PullRequest) ReplaceReviewer(oldReviewerID, newReviewerID string, assignedAt time.Time) ([]PRReviewer, bool) {
	newReviewers := make([]PRReviewer, len(pr.Reviewers))
	replaced := false

	for i, r := range pr.Reviewers {
		if r.UserID == oldReviewerID {
			r.UserID = newReviewerID
			r.AssignedAt = assignedAt
			replaced = true
		}
		newReviewers[i] = r
	}

	return newReviewers, replaced
}

func NormalizeReviewerSlots(reviewers []PRReviewer) {
	for i := range reviewers {
		reviewers[i].Slot = i + 1
	}
}
