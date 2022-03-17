package jobsvc

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
)

// NewJobService creates the job service
func NewJobService(jobRepository repository.JobRepository) JobService {
	return &jobService{
		repo: jobRepository,
	}
}

type jobService struct {
	repo repository.JobRepository
}

func (s jobService) CreateJob(ctx context.Context) (ref.UUID, error) {
	return s.repo.AddJob(ctx, job.Job{})
}

func (s jobService) GetJob(ctx context.Context, ID ref.UUID) (job.Job, error) {
	return s.repo.GetJob(ctx, ID)
}

func (s jobService) ListJobs(ctx context.Context) ([]job.Job, error) {
	return s.repo.ListJobs(ctx)
}
