package jobprocessor

import (
	"testing"
	"time"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/mocks"
	"github.com/KompiTech/itsm-reporting-service/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_processor_ProcessNewJob(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	lastJob := job.Job{}
	err := lastJob.SetUUID("d6aa467b-d07d-41e0-9182-aeedb1b02398")
	require.NoError(t, err)

	lastJob2 := job.Job{}
	err = lastJob2.SetUUID("c0582f65-4c7d-469f-a3a4-42360f287074")
	require.NoError(t, err)

	t.Parallel()

	t.Run("when the processor is busy", func(t *testing.T) {
		repo := new(mocks.JobRepositoryMock)
		repo.Wg.Add(1)
		repo.On("GetLastJob").Return(lastJob, nil).Once()

		jp := NewJobProcessor(logger, repo)
		jp.WaitForJobs()

		err = jp.ProcessNewJob(lastJob.UUID())
		assert.NoError(t, err, "unexpected error", err)

		err = jp.ProcessNewJob(lastJob2.UUID())
		assert.Error(t, err, "expecting error but none returned")

		var domainErr *domain.Error
		assert.ErrorAs(t, err, &domainErr)

		assert.EqualError(t, err, "job is being processed, try it later")

		repo.Wg.Wait()
		repo.AssertExpectations(t)
	})

	t.Run("when the processor is ready", func(t *testing.T) {
		repo := new(mocks.JobRepositoryMock)
		repo.Wg.Add(2)
		repo.On("GetLastJob").Return(lastJob, nil).Once()
		repo.On("GetLastJob").Return(lastJob2, nil).Once()

		jp := NewJobProcessor(logger, repo)
		jp.WaitForJobs()

		err = jp.ProcessNewJob(lastJob.UUID())
		assert.NoError(t, err, "unexpected error", err)
		time.Sleep(time.Millisecond) // wait for processor to get ready - TODO do it better way later
		err = jp.ProcessNewJob(lastJob2.UUID())
		assert.NoError(t, err, "unexpected error", err)

		repo.Wg.Wait()
		repo.AssertExpectations(t)
	})
}
