package repotests

import (
	"context"
	"testing"
	"time"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/mocks"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobRepositoryAddingAndGettingJob(t *testing.T, repo repository.JobRepository, clock repository.Clock) {
	ctx := context.Background()

	job1 := job.Job{Type: job.TypeAll}

	jobID, err := repo.AddJob(ctx, job1)
	require.NoError(t, err)

	nonexistentJobID := ref.UUID("7fca0b71-ffd9-4963-8f04-040faaf4f39c")
	_, err = repo.GetJob(ctx, nonexistentJobID)
	require.Error(t, err)
	require.EqualError(t, err, "error loading job from repository: record was not found")

	retJob, err := repo.GetJob(ctx, jobID)
	require.NoError(t, err)

	assert.Equal(t, jobID, retJob.UUID())
	assert.Empty(t, retJob.ChannelsDownloadFinishedAt)
	assert.Empty(t, retJob.FinalStatus)
	assert.Equal(t, job1.Type, retJob.Type)

	assert.NotEmpty(t, retJob.CreatedAt)
	assert.Equal(t, clock.NowFormatted(), retJob.CreatedAt)
}

func TestJobRepositoryUpdateJob(t *testing.T, repo repository.JobRepository) {
	ctx := context.Background()

	job1 := job.Job{Type: job.TypeAll}

	jobID, err := repo.AddJob(ctx, job1)
	require.NoError(t, err)

	retJob, err := repo.GetJob(ctx, jobID)
	require.NoError(t, err)

	retJob.FinalStatus = "success"
	retJobCreatedAt := retJob.CreatedAt
	retJob.CreatedAt = "some changed value"

	// update job
	retJobID, err := repo.UpdateJob(ctx, retJob)
	require.NoError(t, err)

	assert.Equal(t, jobID, retJobID)

	// get updated job
	updatedJob, err := repo.GetJob(ctx, jobID)
	require.NoError(t, err)

	assert.Equal(t, jobID, updatedJob.UUID())
	assert.Equal(t, "success", updatedJob.FinalStatus)
	assert.Equal(t, retJobCreatedAt, updatedJob.CreatedAt) // this should not be changed
}

func TestJobRepositoryListJobs(t *testing.T, repo repository.JobRepository, clock *mocks.FixedClock) {
	ctx := context.Background()

	job1 := job.Job{Type: job.TypeFE}

	itemsPerPage := 6
	var firstJobID, lastOnTheFirstPageJobID, lastJobID ref.UUID
	for i := 0; i < itemsPerPage+4; i++ {
		clock.AddTime(10 * time.Second)
		jobID, err := repo.AddJob(ctx, job1)
		require.NoError(t, err)
		if i == 0 {
			firstJobID = jobID
		}
		if i == 4 {
			lastOnTheFirstPageJobID = jobID
		}
		lastJobID = jobID
	}

	// 1st page
	pageNum := 0
	retJobs0, err := repo.ListJobs(ctx, uint(pageNum), uint(itemsPerPage))
	require.NoError(t, err)

	assert.Len(t, retJobs0, itemsPerPage)
	// ListJobs returns jobs in reverse order (last one on top)
	assert.Equal(t, lastJobID, retJobs0[0].UUID(), "first job on the 1st page")
	assert.Equal(t, lastOnTheFirstPageJobID, retJobs0[itemsPerPage-1].UUID(), "last job on the 1st page")

	// 2nd page
	pageNum = 1
	retJobs1, err := repo.ListJobs(ctx, uint(pageNum), uint(itemsPerPage))
	require.NoError(t, err)

	assert.Len(t, retJobs1, 4)
	// ListJobs returns jobs in reverse order (last one on top)
	assert.Equal(t, firstJobID, retJobs1[3].UUID(), "last job on the 2nd page")
}

func TestJobRepositoryGetLastJob(t *testing.T, repo repository.JobRepository, clock *mocks.FixedClock) {
	ctx := context.Background()

	_, err := repo.GetLastJob(ctx)
	// there are no jobs yet, it should return error
	require.EqualError(t, err, "no jobs in queue")

	job1 := job.Job{Type: job.TypeAll}
	var lastJobID ref.UUID
	for i := 0; i < 5; i++ {
		clock.AddTime(10 * time.Second)
		lastJobID, err = repo.AddJob(ctx, job1)
		require.NoError(t, err)
	}

	retJob, err := repo.GetLastJob(ctx)
	require.NoError(t, err)

	assert.Equal(t, lastJobID, retJob.UUID())
	assert.Equal(t, job1.Type, retJob.Type)

}
