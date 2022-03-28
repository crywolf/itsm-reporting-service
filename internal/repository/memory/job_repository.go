package memory

import (
	"context"
	"io"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/types"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

const repositorySize = 10

// jobRepositoryMemory keeps data in memory
type jobRepositoryMemory struct {
	Rand  io.Reader
	clock repository.Clock
	jobs  []Job
}

// NewJobRepositoryMemory returns new initialized job repository that keeps data in memory
func NewJobRepositoryMemory(clock repository.Clock) repository.JobRepository {
	return &jobRepositoryMemory{
		clock: clock,
	}
}

// AddJob adds the given job to the repository (repository has fixed length)
func (r *jobRepositoryMemory) AddJob(_ context.Context, _ job.Job) (ref.UUID, error) {
	now := r.clock.NowFormatted().String()

	jobID, err := repository.GenerateUUID(r.Rand)
	if err != nil {
		return ref.UUID(""), err
	}

	storedJob := Job{
		ID:        jobID.String(),
		CreatedAt: now,
	}

	r.jobs = append(r.jobs, storedJob)

	if len(r.jobs) > repositorySize {
		r.jobs = r.jobs[1:] // remove the first element
	}

	return jobID, nil
}

// UpdateJob updates the given job in the repository
func (r *jobRepositoryMemory) UpdateJob(_ context.Context, job job.Job) (ref.UUID, error) {
	storedJob := Job{
		ID:                     job.UUID().String(),
		ProcessingStartedAt:    job.ProcessingStartedAt.String(),
		ChannelsDownloadStatus: job.ChannelsDownloadStatus,
		CreatedAt:              job.CreatedAt.String(),
	}

	for i := range r.jobs {
		if r.jobs[i].ID == job.UUID().String() {
			r.jobs[i] = storedJob
			return job.UUID(), nil
		}
	}

	return job.UUID(), domain.WrapErrorf(ErrNotFound, domain.ErrorCodeNotFound, "error updating job in repository")
}

// GetJob returns the job with the given ID from the repository
func (r jobRepositoryMemory) GetJob(_ context.Context, ID ref.UUID) (job.Job, error) {
	var j job.Job
	var err error

	for i := range r.jobs {
		if r.jobs[i].ID == ID.String() {
			storedJob := r.jobs[i]

			j, err = r.convertStoredToDomainIncident(storedJob)
			if err != nil {
				return job.Job{}, err
			}

			return j, nil
		}
	}

	return job.Job{}, domain.WrapErrorf(ErrNotFound, domain.ErrorCodeNotFound, "error loading job from repository")
}

// GetLastJob returns the last inserted job from the repository
func (r jobRepositoryMemory) GetLastJob(_ context.Context) (job.Job, error) {
	if len(r.jobs) == 0 {
		return job.Job{}, domain.NewErrorf(domain.ErrorCodeUnknown, "no jobs in queue")
	}

	storedJob := r.jobs[len(r.jobs)-1]

	return r.convertStoredToDomainIncident(storedJob)
}

// ListJobs returns the list of jobs from the repository (last one as first)
func (r jobRepositoryMemory) ListJobs(_ context.Context) ([]job.Job, error) {
	var list []job.Job

	for i := len(r.jobs) - 1; i >= 0; i-- {
		storedJob := r.jobs[i]
		j, err := r.convertStoredToDomainIncident(storedJob)
		if err != nil {
			return list, err
		}

		list = append(list, j)
	}

	return list, nil
}

func (r jobRepositoryMemory) convertStoredToDomainIncident(storedJob Job) (job.Job, error) {
	var j job.Job
	errMsg := "error loading job from repository (%s)"

	err := j.SetUUID(ref.UUID(storedJob.ID))
	if err != nil {
		return job.Job{}, domain.WrapErrorf(err, domain.ErrorCodeUnknown, errMsg, "storedJob.ID")
	}

	j.CreatedAt = types.DateTime(storedJob.CreatedAt)
	j.ProcessingStartedAt = types.DateTime(storedJob.ProcessingStartedAt)
	j.ChannelsDownloadStatus = storedJob.ChannelsDownloadStatus

	return j, nil
}
