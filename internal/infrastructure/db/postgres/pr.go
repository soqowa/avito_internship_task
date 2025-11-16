package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domain "github.com/user/reviewer-svc/internal/domain"
	domainpr "github.com/user/reviewer-svc/internal/domain/pr"
	stats "github.com/user/reviewer-svc/internal/domain/stats"
	userreassign "github.com/user/reviewer-svc/internal/domain/userreassign"
)

const (
	statusOpenSmallint   int16 = 1
	statusMergedSmallint int16 = 2
)


type PRRepo struct{}

func NewPRRepo() *PRRepo {
	return &PRRepo{}
}

var (
	_ domainpr.PullRequestRepository              = (*PRRepo)(nil)
	_ stats.PullRequestStatsRepository            = (*PRRepo)(nil)
	_ userreassign.ReassignmentPRRepository       = (*PRRepo)(nil)
)

func (r *PRRepo) Create(ctx context.Context, ttx domain.Tx, pr *domainpr.PullRequest) error {
	_, err := ttx.Exec(ctx,
		"INSERT INTO pull_requests (id, title, author_id, status, created_at, merged_at) VALUES ($1, $2, $3, $4, $5, $6)",
		pr.ID, pr.Title, pr.AuthorID, statusToSmallint(pr.Status), pr.CreatedAt, pr.MergedAt,
	)
	if err != nil {
		return translateError(err)
	}
	for _, rv := range pr.Reviewers {
		_, err := ttx.Exec(ctx,
			"INSERT INTO pr_reviewers (pr_id, slot, user_id, created_at) VALUES ($1, $2, $3, $4)",
			pr.ID, rv.Slot, rv.UserID, rv.AssignedAt,
		)
		if err != nil {
			return translateError(err)
		}
	}
	return nil
}

func (r *PRRepo) GetByID(ctx context.Context, ttx domain.Tx, id uuid.UUID, forUpdate bool) (*domainpr.PullRequest, error) {
	query := "SELECT id, title, author_id, status, created_at, merged_at FROM pull_requests WHERE id = $1"
	if forUpdate {
		query += " FOR UPDATE"
	}
	row := ttx.QueryRow(ctx, query, id)
	var pr domainpr.PullRequest
	var statusSmall int16
	if err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &statusSmall, &pr.CreatedAt, &pr.MergedAt); err != nil {
		return nil, translateError(err)
	}
	pr.Status = statusFromSmallint(statusSmall)

	reviewers, err := r.loadReviewers(ctx, ttx, pr.ID)
	if err != nil {
		return nil, err
	}
	pr.Reviewers = reviewers
	return &pr, nil
}

func (r *PRRepo) UpdateStatus(ctx context.Context, ttx domain.Tx, id uuid.UUID, status domainpr.PRStatus, mergedAt *time.Time) error {
	_, err := ttx.Exec(ctx,
		"UPDATE pull_requests SET status = $1, merged_at = $2 WHERE id = $3",
		statusToSmallint(status), mergedAt, id,
	)
	return translateError(err)
}

func (r *PRRepo) ReplaceReviewers(ctx context.Context, ttx domain.Tx, prID uuid.UUID, reviewers []domainpr.PRReviewer) error {
	if _, err := ttx.Exec(ctx, "DELETE FROM pr_reviewers WHERE pr_id = $1", prID); err != nil {
		return translateError(err)
	}
	for _, rv := range reviewers {
		_, err := ttx.Exec(ctx,
			"INSERT INTO pr_reviewers (pr_id, slot, user_id, created_at) VALUES ($1, $2, $3, $4)",
			prID, rv.Slot, rv.UserID, rv.AssignedAt,
		)
		if err != nil {
			return translateError(err)
		}
	}
	return nil
}

func (r *PRRepo) List(ctx context.Context, ttx domain.Tx, status *domainpr.PRStatus) ([]domainpr.PullRequest, error) {
	query := "SELECT id, title, author_id, status, created_at, merged_at FROM pull_requests"
	var args []any
	if status != nil {
		query += " WHERE status = $1"
		args = append(args, statusToSmallint(*status))
	}
	query += " ORDER BY created_at DESC"

	rows, err := ttx.Query(ctx, query, args...)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	var res []domainpr.PullRequest
	var ids []uuid.UUID
	for rows.Next() {
		var pr domainpr.PullRequest
		var statusSmall int16
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &statusSmall, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, err
		}
		pr.Status = statusFromSmallint(statusSmall)
		ids = append(ids, pr.ID)
		res = append(res, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	reviewersByPR, err := r.loadReviewersBulk(ctx, ttx, ids)
	if err != nil {
		return nil, err
	}
	for i := range res {
		if rv, ok := reviewersByPR[res[i].ID]; ok {
			res[i].Reviewers = rv
		}
	}
	return res, nil
}

func (r *PRRepo) ListAssignedTo(ctx context.Context, ttx domain.Tx, userID uuid.UUID, status *domainpr.PRStatus) ([]domainpr.PullRequest, error) {
	query := "SELECT DISTINCT p.id, p.title, p.author_id, p.status, p.created_at, p.merged_at FROM pull_requests p JOIN pr_reviewers r ON r.pr_id = p.id WHERE r.user_id = $1"
	args := []any{userID}
	if status != nil {
		query += " AND p.status = $2"
		args = append(args, statusToSmallint(*status))
	}
	query += " ORDER BY p.created_at DESC"

	rows, err := ttx.Query(ctx, query, args...)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	var res []domainpr.PullRequest
	var ids []uuid.UUID
	for rows.Next() {
		var pr domainpr.PullRequest
		var statusSmall int16
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &statusSmall, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, err
		}
		pr.Status = statusFromSmallint(statusSmall)
		ids = append(ids, pr.ID)
		res = append(res, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	reviewersByPR, err := r.loadReviewersBulk(ctx, ttx, ids)
	if err != nil {
		return nil, err
	}
	for i := range res {
		if rv, ok := reviewersByPR[res[i].ID]; ok {
			res[i].Reviewers = rv
		}
	}
	return res, nil
}

func (r *PRRepo) StatsByUser(ctx context.Context, ttx domain.Tx, teamID *uuid.UUID) ([]stats.UserAssignmentsStats, error) {
	query := "SELECT u.id, COUNT(prr.pr_id) AS total," +
		" COUNT(CASE WHEN p.status = $1 THEN 1 END) AS open_cnt," +
		" COUNT(CASE WHEN p.status = $2 THEN 1 END) AS merged_cnt" +
		" FROM users u" +
		" LEFT JOIN pr_reviewers prr ON prr.user_id = u.id" +
		" LEFT JOIN pull_requests p ON p.id = prr.pr_id"
	args := []any{statusOpenSmallint, statusMergedSmallint}
	if teamID != nil {
		args = append(args, *teamID)
		query += fmt.Sprintf(" WHERE u.team_id = $%d", len(args))
	}
	query += " GROUP BY u.id ORDER BY u.id"

	rows, err := ttx.Query(ctx, query, args...)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	var res []stats.UserAssignmentsStats
	for rows.Next() {
		var s stats.UserAssignmentsStats
		if err := rows.Scan(&s.UserID, &s.TotalAssigned, &s.OpenAssigned, &s.MergedAssigned); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *PRRepo) StatsByPR(ctx context.Context, ttx domain.Tx, teamID *uuid.UUID) ([]stats.PRAssignmentsStats, error) {
	query := "SELECT p.id, COUNT(prr.user_id) AS reviewers_cnt FROM pull_requests p"
	var args []any
	if teamID != nil {
		query += " JOIN users u ON p.author_id = u.id AND u.team_id = $1"
		args = append(args, *teamID)
	}
	query += " LEFT JOIN pr_reviewers prr ON prr.pr_id = p.id"
	query += " GROUP BY p.id ORDER BY p.id"

	rows, err := ttx.Query(ctx, query, args...)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	var res []stats.PRAssignmentsStats
	for rows.Next() {
		var s stats.PRAssignmentsStats
		if err := rows.Scan(&s.PRID, &s.ReviewersCount); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *PRRepo) loadReviewers(ctx context.Context, ttx domain.Tx, prID uuid.UUID) ([]domainpr.PRReviewer, error) {
	rows, err := ttx.Query(ctx,
		"SELECT pr_id, slot, user_id, created_at FROM pr_reviewers WHERE pr_id = $1 ORDER BY slot",
		prID,
	)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	var res []domainpr.PRReviewer
	for rows.Next() {
		var rv domainpr.PRReviewer
		if err := rows.Scan(&rv.PRID, &rv.Slot, &rv.UserID, &rv.AssignedAt); err != nil {
			return nil, err
		}
		res = append(res, rv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *PRRepo) loadReviewersBulk(ctx context.Context, ttx domain.Tx, prIDs []uuid.UUID) (map[uuid.UUID][]domainpr.PRReviewer, error) {
	if len(prIDs) == 0 {
		return map[uuid.UUID][]domainpr.PRReviewer{}, nil
	}

	query, args := buildUUIDInQuery(
		"SELECT pr_id, slot, user_id, created_at FROM pr_reviewers WHERE pr_id IN (",
		") ORDER BY pr_id, slot",
		prIDs,
	)

	rows, err := ttx.Query(ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	res := make(map[uuid.UUID][]domainpr.PRReviewer)
	for rows.Next() {
		var rv domainpr.PRReviewer
		if err := rows.Scan(&rv.PRID, &rv.Slot, &rv.UserID, &rv.AssignedAt); err != nil {
			return nil, err
		}
		res[rv.PRID] = append(res[rv.PRID], rv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func statusToSmallint(s domainpr.PRStatus) int16 {
	switch s {
	case domainpr.PRStatusOpen:
		return statusOpenSmallint
	case domainpr.PRStatusMerged:
		return statusMergedSmallint
	default:
		return statusOpenSmallint
	}
}

func statusFromSmallint(v int16) domainpr.PRStatus {
	switch v {
	case statusOpenSmallint:
		return domainpr.PRStatusOpen
	case statusMergedSmallint:
		return domainpr.PRStatusMerged
	default:
		return domainpr.PRStatusOpen
	}
}

var _ domainpr.PullRequestRepository = (*PRRepo)(nil)
var _ stats.PullRequestStatsRepository = (*PRRepo)(nil)
