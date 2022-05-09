package sql

import (
	"database/sql"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/KompiTech/itsm-reporting-service/internal/mocks"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
	repotests "github.com/KompiTech/itsm-reporting-service/internal/repository/tests"
	"github.com/cockroachdb/copyist"
	"github.com/stretchr/testify/require"
)

func init() {
	copyist.Register("postgres")
}

// DB is a shared database handle
var DB *sql.DB

func resetDB(db *sql.DB) {
	if _, err := db.Exec("TRUNCATE jobs"); err != nil {
		panic(err)
	}
}

func newJobRepositorySQL(t *testing.T) (repository.JobRepository, *mocks.FixedClock) {
	connStr := os.Getenv("TEST_DB_CONNECTION_STRING") // postgresql://root@localhost:26257?sslmode=disable

	if DB == nil {
		var err error
		DB, err = sql.Open("copyist_postgres", connStr)
		if err != nil {
			panic(err)
		}
	}

	clock := mocks.NewFixedClock()

	// deterministic "random number generator" to generate deterministic UUIDs in tests
	rand := strings.NewReader(
		"XVlBzgbaiCMRAjWwhTHctcuAxhxKQFDaFpLSjFbcXoEFfRsWxPLDnJObCsNVlgTeMaPEZQleQYhYzRyWJjPjzpfRFEgmotaFetHsbZRjxAw" +
			"nwekrBEmfdzdcEkXBAkjQZLCtTMtTCoaNatyyiNKAReKJyiXJrscctNswYNsGRussVmaozFZBsbOJiFQGZsnwTKSmVoiG",
	)

	repo, err := NewJobRepositorySQL(clock, DB, rand)
	require.NoError(t, err)

	resetDB(DB)

	return repo, clock
}

func TestJobRepositorySQL_AddingAndGettingJob(t *testing.T) {
	defer func(open io.Closer) { _ = open.Close() }(copyist.Open(t))

	repo, clock := newJobRepositorySQL(t)
	repotests.TestJobRepositoryAddingAndGettingJob(t, repo, clock)
}

func TestJobRepositorySQL_UpdateJob(t *testing.T) {
	defer func(open io.Closer) { _ = open.Close() }(copyist.Open(t))

	repo, _ := newJobRepositorySQL(t)
	repotests.TestJobRepositoryUpdateJob(t, repo)
}

func TestJobRepositorySQL_ListJobs(t *testing.T) {
	defer func(open io.Closer) { _ = open.Close() }(copyist.Open(t))

	repo, clock := newJobRepositorySQL(t)
	repotests.TestJobRepositoryListJobs(t, repo, clock, repositorySize)
}

func TestJobRepositorySQL_GetLastJob(t *testing.T) {
	defer func(open io.Closer) { _ = open.Close() }(copyist.Open(t))

	repo, clock := newJobRepositorySQL(t)
	repotests.TestJobRepositoryGetLastJob(t, repo, clock)
}
