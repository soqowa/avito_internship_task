package stats

import "github.com/google/uuid"

type UserAssignmentsStats struct {
	UserID         uuid.UUID
	TotalAssigned  int
	OpenAssigned   int
	MergedAssigned int
}

type PRAssignmentsStats struct {
	PRID           uuid.UUID
	ReviewersCount int
}
