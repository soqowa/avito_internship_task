package stats

import "github.com/google/uuid"

type UserAssignmentsStatsItem struct {
	UserID         uuid.UUID `json:"userId"`
	TotalAssigned  int       `json:"totalAssigned"`
	OpenAssigned   int       `json:"openAssigned"`
	MergedAssigned int       `json:"mergedAssigned"`
}

type UserAssignmentsStatsResponse struct {
	Items []UserAssignmentsStatsItem `json:"items"`
}

type PRAssignmentsStatsItem struct {
	PRID           uuid.UUID `json:"prId"`
	ReviewersCount int       `json:"reviewersCount"`
}

type PRAssignmentsStatsResponse struct {
	Items []PRAssignmentsStatsItem `json:"items"`
}
