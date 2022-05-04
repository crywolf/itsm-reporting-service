package memory

import (
	"testing"

	"github.com/KompiTech/itsm-reporting-service/internal/mocks"
	repotests "github.com/KompiTech/itsm-reporting-service/internal/repository/tests"
)

func TestJobRepositoryMemory_AddingAndGettingJob(t *testing.T) {
	clock := mocks.NewFixedClock()
	repo := NewJobRepositoryMemory(clock)

	repotests.TestJobRepositoryAddingAndGettingJob(t, repo, clock)
}

func TestJobRepositoryMemory_UpdateJob(t *testing.T) {
	clock := mocks.NewFixedClock()
	repo := NewJobRepositoryMemory(clock)

	repotests.TestJobRepositoryUpdateJob(t, repo)
}

func TestJobRepositoryMemory_ListJobs(t *testing.T) {
	clock := mocks.NewFixedClock()
	repo := NewJobRepositoryMemory(clock)

	repotests.TestJobRepositoryListJobs(t, repo, clock, repositorySize)
}

func TestJobRepositoryMemory_GetLastJob(t *testing.T) {
	clock := mocks.NewFixedClock()
	repo := NewJobRepositoryMemory(clock)

	repotests.TestJobRepositoryGetLastJob(t, repo, clock)
}
