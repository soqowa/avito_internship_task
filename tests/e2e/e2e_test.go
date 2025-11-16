package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/user/reviewer-svc/internal/app"
	postgresAdapter "github.com/user/reviewer-svc/internal/infrastructure/db/postgres"
	"github.com/user/reviewer-svc/internal/infrastructure/logger"
)

type e2eTeam struct {
	TeamName string        `json:"team_name"`
	Members  []e2eMember   `json:"members"`
}

type e2eMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type e2eTeamResponse struct {
	Team e2eTeam `json:"team"`
}

type e2ePullRequest struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         *string  `json:"createdAt,omitempty"`
	MergedAt          *string  `json:"mergedAt,omitempty"`
}

type e2ePRResponse struct {
	PR e2ePullRequest `json:"pr"`
}

func setupApp(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := tcpostgres.RunContainer(ctx,
		tcpostgres.WithDatabase("reviewer"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		t.Fatalf("failed to get dsn: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		t.Fatalf("getwd: %v", err)
	}
	migrationsDir := filepath.Join(wd, "..", "..", "migrations")

	var migrateErr error
	for i := 0; i < 5; i++ {
		migrateErr = postgresAdapter.Migrate(dsn, migrationsDir)
		if migrateErr == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if migrateErr != nil {
		_ = pgContainer.Terminate(ctx)
		t.Fatalf("migrate: %v", migrateErr)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		t.Fatalf("pgxpool: %v", err)
	}

	lg := logger.New("debug")
	r := chi.NewRouter()
	handler := app.NewHandler(r, pool, lg)

	ts := httptest.NewServer(handler)

	cleanup := func() {
		pool.Close()
		ts.Close()
		_ = pgContainer.Terminate(context.Background())
	}

	return ts, cleanup
}

func TestPRCreateAndMergeIdempotent(t *testing.T) {
	ts, cleanup := setupApp(t)
	defer cleanup()

	client := &http.Client{Timeout: 5 * time.Second}

	teamPayload := `{
		"team_name": "team-e2e",
		"members": [
			{"user_id": "u1", "username": "author", "is_active": true},
			{"user_id": "u2", "username": "reviewer1", "is_active": true},
			{"user_id": "u3", "username": "reviewer2", "is_active": true}
		]
	}`

	teamRes, err := client.Post(ts.URL+"/team/add", "application/json", strings.NewReader(teamPayload))
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	defer teamRes.Body.Close()
	if teamRes.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", teamRes.StatusCode)
	}
	var teamResp e2eTeamResponse
	if err := json.NewDecoder(teamRes.Body).Decode(&teamResp); err != nil {
		t.Fatalf("decode team: %v", err)
	}

	prBody := `{"pull_request_id": "pr-1001", "pull_request_name": "pr1", "author_id": "u1"}`
	prRes, err := client.Post(ts.URL+"/pullRequest/create", "application/json", strings.NewReader(prBody))
	if err != nil {
		t.Fatalf("create pr: %v", err)
	}
	defer prRes.Body.Close()
	if prRes.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", prRes.StatusCode)
	}
	var prResp e2ePRResponse
	if err := json.NewDecoder(prRes.Body).Decode(&prResp); err != nil {
		t.Fatalf("decode pr: %v", err)
	}
	if prResp.PR.Status != "OPEN" {
		t.Fatalf("expected status OPEN, got %s", prResp.PR.Status)
	}

	mergeBody := `{"pull_request_id": "pr-1001"}`
	for i := 0; i < 2; i++ {
		mergeRes, err := client.Post(ts.URL+"/pullRequest/merge", "application/json", strings.NewReader(mergeBody))
		if err != nil {
			t.Fatalf("merge pr: %v", err)
		}
		if mergeRes.StatusCode != http.StatusOK {
			mergeRes.Body.Close()
			t.Fatalf("expected 200, got %d", mergeRes.StatusCode)
		}
		var mergedResp e2ePRResponse
		if err := json.NewDecoder(mergeRes.Body).Decode(&mergedResp); err != nil {
			mergeRes.Body.Close()
			t.Fatalf("decode merged pr: %v", err)
		}
		mergeRes.Body.Close()
		if mergedResp.PR.Status != "MERGED" {
			t.Fatalf("expected MERGED, got %s", mergedResp.PR.Status)
		}
		if mergedResp.PR.MergedAt == nil {
			t.Fatalf("expected mergedAt to be set")
		}
	}
}
