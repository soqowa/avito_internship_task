package stats

type UserAssignmentsStats struct {
	UserID         string
	TotalAssigned  int
	OpenAssigned   int
	MergedAssigned int
}

type PRAssignmentsStats struct {
	PRID           string
	ReviewersCount int
}
