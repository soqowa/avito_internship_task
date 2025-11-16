package prs

import (
	"time"

	"github.com/google/uuid"
)

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PRReviewer struct {
	UserID     uuid.UUID `json:"userId"`
	Slot       int       `json:"slot"`
	AssignedAt time.Time `json:"assignedAt"`
}

type PullRequest struct {
	ID        uuid.UUID    `json:"id"`
	Title     string       `json:"title"`
	AuthorID  uuid.UUID    `json:"authorId"`
	Status    PRStatus     `json:"status"`
	Reviewers []PRReviewer `json:"reviewers"`
	CreatedAt time.Time    `json:"createdAt"`
	MergedAt  *time.Time   `json:"mergedAt"`
}

type CreatePRRequest struct {
	Title    string    `json:"title"`
	AuthorID uuid.UUID `json:"authorId"`
}

type ReassignReviewerRequest struct {
	OldReviewerID uuid.UUID `json:"oldReviewerId"`
}
