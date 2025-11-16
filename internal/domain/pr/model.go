package pr

import (
	"time"

	"github.com/google/uuid"
)

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PRReviewer struct {
	PRID       uuid.UUID
	Slot       int
	UserID     uuid.UUID
	AssignedAt time.Time
}

type PullRequest struct {
	ID        uuid.UUID
	Title     string
	AuthorID  uuid.UUID
	Status    PRStatus
	CreatedAt time.Time
	MergedAt  *time.Time
	Reviewers []PRReviewer
}
