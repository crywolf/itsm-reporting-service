package jobsvc

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/api"
	converters "github.com/KompiTech/itsm-reporting-service/internal/http/rest/api/input_converters"
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

func (s jobService) CreateJob(ctx context.Context, params api.CreateJobParams) (ref.UUID, error) {
	return s.repo.AddJob(ctx, job.Job{
		Type: params.Type,
	})
}

func (s jobService) UpdateJob(ctx context.Context, j job.Job) (ref.UUID, error) {
	return s.repo.UpdateJob(ctx, j)
}

func (s jobService) GetJob(ctx context.Context, ID ref.UUID) (job.Job, error) {
	return s.repo.GetJob(ctx, ID)
}

func (s jobService) ListJobs(ctx context.Context, paginationParams converters.PaginationParams) ([]job.Job, error) {
	return s.repo.ListJobs(ctx, paginationParams.Page(), paginationParams.ItemsPerPage())
}
