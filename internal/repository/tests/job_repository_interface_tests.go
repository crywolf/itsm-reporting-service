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

	job1 := job.Job{}

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

	assert.NotEmpty(t, retJob.CreatedAt)
	assert.Equal(t, clock.NowFormatted(), retJob.CreatedAt)
}

func TestJobRepositoryUpdateJob(t *testing.T, repo repository.JobRepository) {
	ctx := context.Background()

	job1 := job.Job{}

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

func TestJobRepositoryListJobs(t *testing.T, repo repository.JobRepository, clock *mocks.FixedClock, repositorySize int) {
	ctx := context.Background()

	job1 := job.Job{}

	var thirdJobID, lastJobID ref.UUID
	for i := 0; i < repositorySize+2; i++ {
		clock.AddTime(10 * time.Second)
		jobID, err := repo.AddJob(ctx, job1)
		if i == 2 {
			thirdJobID = jobID
		}
		lastJobID = jobID
		require.NoError(t, err)
	}

	retJobs, err := repo.ListJobs(ctx)
	require.NoError(t, err)

	// repo has fixed size first two jobs are discarded (FIFO)
	assert.Len(t, retJobs, repositorySize)
	// ListJobs returns jobs in reverse order (last one on top, 3rd would be last)
	assert.Equal(t, lastJobID, retJobs[0].UUID(), "last job")
	assert.Equal(t, thirdJobID, retJobs[repositorySize-1].UUID(), "third job")
}

func TestJobRepositoryGetLastJob(t *testing.T, repo repository.JobRepository, clock *mocks.FixedClock) {
	ctx := context.Background()

	_, err := repo.GetLastJob(ctx)
	// there are no jobs yet, it should return error
	require.EqualError(t, err, "no jobs in queue")

	job1 := job.Job{}
	var lastJobID ref.UUID
	for i := 0; i < 5; i++ {
		clock.AddTime(10 * time.Second)
		lastJobID, err = repo.AddJob(ctx, job1)
		require.NoError(t, err)
	}

	retJob, err := repo.GetLastJob(ctx)
	require.NoError(t, err)

	assert.Equal(t, lastJobID, retJob.UUID())
}
