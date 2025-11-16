package stats

import dstats "github.com/user/reviewer-svc/internal/domain/stats"

func toResponse(v interface{}) interface{} {
	switch s := v.(type) {
	case dstats.UserAssignmentsStats:
		return UserAssignmentsStatsItem{
			UserID:         s.UserID,
			TotalAssigned:  s.TotalAssigned,
			OpenAssigned:   s.OpenAssigned,
			MergedAssigned: s.MergedAssigned,
		}
	case dstats.PRAssignmentsStats:
		return PRAssignmentsStatsItem{
			PRID:           s.PRID,
			ReviewersCount: s.ReviewersCount,
		}
	default:
		return nil
	}
}
