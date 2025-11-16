package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/user/reviewer-svc/internal/app"
	postgresAdapter "github.com/user/reviewer-svc/internal/infrastructure/db/postgres"
	"github.com/user/reviewer-svc/internal/infrastructure/logger"
)

type e2eTeam struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type e2eUser struct {
	Id uuid.UUID `json:"id"`
}

type e2ePullRequest struct {
	Id       uuid.UUID  `json:"id"`
	Status   string     `json:"status"`
	MergedAt *time.Time `json:"mergedAt"`
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

	teamRes, err := client.Post(ts.URL+"/teams", "application/json", strings.NewReader(`{"name":"team-e2e"}`))
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	defer teamRes.Body.Close()
	if teamRes.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", teamRes.StatusCode)
	}
	var team e2eTeam
	if err := json.NewDecoder(teamRes.Body).Decode(&team); err != nil {
		t.Fatalf("decode team: %v", err)
	}

	userRes, err := client.Post(ts.URL+"/teams/"+team.Id.String()+"/users", "application/json", strings.NewReader(`{"name":"author"}`))
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	defer userRes.Body.Close()
	if userRes.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", userRes.StatusCode)
	}
	var user e2eUser
	if err := json.NewDecoder(userRes.Body).Decode(&user); err != nil {
		t.Fatalf("decode user: %v", err)
	}

	prBody := fmt.Sprintf(`{"title":"pr1","authorId":"%s"}`, user.Id.String())
	prRes, err := client.Post(ts.URL+"/prs", "application/json", strings.NewReader(prBody))
	if err != nil {
		t.Fatalf("create pr: %v", err)
	}
	defer prRes.Body.Close()
	if prRes.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", prRes.StatusCode)
	}
	var pr e2ePullRequest
	if err := json.NewDecoder(prRes.Body).Decode(&pr); err != nil {
		t.Fatalf("decode pr: %v", err)
	}
	if pr.Status != "OPEN" {
		t.Fatalf("expected status OPEN, got %s", pr.Status)
	}

	mergeURL := ts.URL + "/prs/" + pr.Id.String() + "/merge"
	for i := 0; i < 2; i++ {
		mergeRes, err := client.Post(mergeURL, "application/json", nil)
		if err != nil {
			t.Fatalf("merge pr: %v", err)
		}
		if mergeRes.StatusCode != http.StatusOK {
			mergeRes.Body.Close()
			t.Fatalf("expected 200, got %d", mergeRes.StatusCode)
		}
		var merged e2ePullRequest
		if err := json.NewDecoder(mergeRes.Body).Decode(&merged); err != nil {
			mergeRes.Body.Close()
			t.Fatalf("decode merged pr: %v", err)
		}
		mergeRes.Body.Close()
		if merged.Status != "MERGED" {
			t.Fatalf("expected MERGED, got %s", merged.Status)
		}
		if merged.MergedAt == nil {
			t.Fatalf("expected mergedAt to be set")
		}
	}
}
