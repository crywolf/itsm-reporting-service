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

// AddJob adds the given job to the repository
func (r *jobRepositoryMemory) AddJob(_ context.Context, job job.Job) (ref.UUID, error) {
	now := r.clock.NowFormatted().String()

	jobID, err := repository.GenerateUUID(r.Rand)
	if err != nil {
		return ref.UUID(""), err
	}

	storedJob := Job{
		ID:        jobID.String(),
		Type:      job.Type.String(),
		CreatedAt: now,
	}

	r.jobs = append(r.jobs, storedJob)

	return jobID, nil
}

// UpdateJob updates the given job in the repository
func (r *jobRepositoryMemory) UpdateJob(_ context.Context, job job.Job) (ref.UUID, error) {
	storedJob := Job{
		ID:                             job.UUID().String(),
		Type:                           job.Type.String(),
		ChannelsDownloadStartedAt:      job.ChannelsDownloadStartedAt.String(),
		ChannelsDownloadFinishedAt:     job.ChannelsDownloadFinishedAt.String(),
		UsersDownloadStartedAt:         job.UsersDownloadStartedAt.String(),
		UsersDownloadFinishedAt:        job.UsersDownloadFinishedAt.String(),
		TicketsDownloadStartedAt:       job.TicketsDownloadStartedAt.String(),
		TicketsDownloadFinishedAt:      job.TicketsDownloadFinishedAt.String(),
		ExcelFilesGenerationStartedAt:  job.ExcelFilesGenerationStartedAt.String(),
		ExcelFilesGenerationFinishedAt: job.ExcelFilesGenerationFinishedAt.String(),
		EmailsSendingStartedAt:         job.EmailsSendingStartedAt.String(),
		EmailsSendingFinishedAt:        job.EmailsSendingFinishedAt.String(),
		FinalStatus:                    job.FinalStatus,
	}

	for i, origJob := range r.jobs {
		if r.jobs[i].ID == job.UUID().String() {
			storedJob.CreatedAt = origJob.CreatedAt // this cannot be changed

			r.jobs[i] = storedJob
			return job.UUID(), nil
		}
	}

	return job.UUID(), domain.WrapErrorf(repository.ErrNotFound, domain.ErrorCodeNotFound, "error updating job in repository")
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

	return job.Job{}, domain.WrapErrorf(repository.ErrNotFound, domain.ErrorCodeNotFound, "error loading job from repository")
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
func (r jobRepositoryMemory) ListJobs(_ context.Context, page, perPage uint) ([]job.Job, error) {
	var list []job.Job

	total := uint(len(r.jobs))

	start := page * perPage
	if start >= total {
		start = total
	}

	end := start + perPage
	if end > total {
		end = total
	}

	lastIndex := int(total - start - 1)
	firstIndex := int(total - end)

	for i := lastIndex; i >= firstIndex; i-- {
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
		return j, domain.WrapErrorf(err, domain.ErrorCodeUnknown, errMsg, "storedJob.ID")
	}

	j.Type, err = job.NewTypeFromString(storedJob.Type)
	if err != nil {
		return j, domain.WrapErrorf(err, domain.ErrorCodeUnknown, errMsg, "storedJob.Type")
	}
	j.CreatedAt = types.DateTime(storedJob.CreatedAt)
	j.ChannelsDownloadStartedAt = types.DateTime(storedJob.ChannelsDownloadStartedAt)
	j.ChannelsDownloadFinishedAt = types.DateTime(storedJob.ChannelsDownloadFinishedAt)
	j.UsersDownloadStartedAt = types.DateTime(storedJob.UsersDownloadStartedAt)
	j.UsersDownloadFinishedAt = types.DateTime(storedJob.UsersDownloadFinishedAt)
	j.TicketsDownloadStartedAt = types.DateTime(storedJob.TicketsDownloadStartedAt)
	j.TicketsDownloadFinishedAt = types.DateTime(storedJob.TicketsDownloadFinishedAt)
	j.ExcelFilesGenerationStartedAt = types.DateTime(storedJob.ExcelFilesGenerationStartedAt)
	j.ExcelFilesGenerationFinishedAt = types.DateTime(storedJob.ExcelFilesGenerationFinishedAt)
	j.EmailsSendingStartedAt = types.DateTime(storedJob.EmailsSendingStartedAt)
	j.EmailsSendingFinishedAt = types.DateTime(storedJob.EmailsSendingFinishedAt)
	j.FinalStatus = storedJob.FinalStatus

	return j, nil
}
