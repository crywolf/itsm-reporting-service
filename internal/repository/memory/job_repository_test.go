package memory

import (
	"context"
	"testing"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobRepositoryMemory_AddingAndGettingJob(t *testing.T) {
	clock := mocks.NewFixedClock()
	ctx := context.Background()

	repo := NewJobRepositoryMemory(clock)

	job1 := job.Job{}

	jobID, err := repo.AddJob(ctx, job1)
	require.NoError(t, err)

	retJob, err := repo.GetJob(ctx, jobID)
	require.NoError(t, err)

	assert.Equal(t, jobID, retJob.UUID())
	assert.Empty(t, retJob.ProcessingStartedAt)
	assert.Empty(t, retJob.ChannelsDownloadStatus)

	assert.NotEmpty(t, retJob.CreatedAt)
	assert.Equal(t, clock.NowFormatted(), retJob.CreatedAt)
}

func TestJobRepositoryMemory_UpdateJob(t *testing.T) {
	clock := mocks.NewFixedClock()
	ctx := context.Background()

	repo := NewJobRepositoryMemory(clock)

	job1 := job.Job{}

	jobID, err := repo.AddJob(ctx, job1)
	require.NoError(t, err)

	retJob, err := repo.GetJob(ctx, jobID)
	require.NoError(t, err)

	retJob.ChannelsDownloadStatus = "success"

	// update job
	retJobID, err := repo.UpdateJob(ctx, retJob)
	require.NoError(t, err)

	assert.Equal(t, jobID, retJobID)

	// get updated job
	updatedJob, err := repo.GetJob(ctx, jobID)
	require.NoError(t, err)

	assert.Equal(t, jobID, updatedJob.UUID())
	assert.Equal(t, "success", updatedJob.ChannelsDownloadStatus)
	assert.Equal(t, retJob.CreatedAt, updatedJob.CreatedAt) // this should not be changed
}

func TestJobRepositoryMemory_ListJob(t *testing.T) {
	clock := mocks.NewFixedClock()
	ctx := context.Background()

	repo := NewJobRepositoryMemory(clock)

	job1 := job.Job{}

	var thirdJobID, lastJobID ref.UUID
	for i := 0; i < repositorySize+2; i++ {
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

func TestJobRepositoryMemory_GetLastJob(t *testing.T) {
	clock := mocks.NewFixedClock()
	ctx := context.Background()

	repo := NewJobRepositoryMemory(clock)

	_, err := repo.GetLastJob(ctx)
	// there are no jobs yet, it should return error
	require.EqualError(t, err, "no jobs in queue")

	job1 := job.Job{}
	var lastJobID ref.UUID
	for i := 0; i < 5; i++ {
		lastJobID, err = repo.AddJob(ctx, job1)
		require.NoError(t, err)
	}

	retJob, err := repo.GetLastJob(ctx)
	require.NoError(t, err)

	assert.Equal(t, lastJobID, retJob.UUID())
}
