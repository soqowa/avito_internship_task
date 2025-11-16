package stats

type UserAssignmentsStatsItem struct {
	UserID         string `json:"userId"`
	TotalAssigned  int    `json:"totalAssigned"`
	OpenAssigned   int    `json:"openAssigned"`
	MergedAssigned int    `json:"mergedAssigned"`
}

type UserAssignmentsStatsResponse struct {
	Items []UserAssignmentsStatsItem `json:"items"`
}

type PRAssignmentsStatsItem struct {
	PRID           string `json:"prId"`
	ReviewersCount int    `json:"reviewersCount"`
}

type PRAssignmentsStatsResponse struct {
	Items []PRAssignmentsStatsItem `json:"items"`
}
