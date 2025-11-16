package domain

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyMerged = errors.New("already merged")
	ErrNoCandidate   = errors.New("no candidate")
	ErrBadReviewer   = errors.New("bad reviewer")
	ErrAlreadyExists = errors.New("already exists")

	ErrInvalidRequest = errors.New("invalid request")

	ErrInvalidTeamName   = errors.New("invalid team name")
	ErrInvalidUserName   = errors.New("invalid user name")
	ErrInvalidPRTitle    = errors.New("invalid PR title")
	ErrEmptyUpdate       = errors.New("no fields to update")
	ErrEmptyBulkUserIDs  = errors.New("empty bulk user IDs")
	ErrCrossTeamDeactive = errors.New("user does not belong to team")

	ErrConstraintViolation = errors.New("constraint violation")
)
