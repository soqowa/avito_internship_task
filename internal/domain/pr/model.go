package pr

import (
	"time"
)

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PRReviewer struct {
	PRID       string
	Slot       int
	UserID     string
	AssignedAt time.Time
}

type PullRequest struct {
	ID        string
	Title     string
	AuthorID  string
	Status    PRStatus
	CreatedAt time.Time
	MergedAt  *time.Time
	Reviewers []PRReviewer
}
