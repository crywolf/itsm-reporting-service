package jobsvc

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/api"
	converters "github.com/KompiTech/itsm-reporting-service/internal/http/rest/api/input_converters"
)

// JobService provides job operations
type JobService interface {
	// CreateJob creates new job and adds it to the repository
	CreateJob(ctx context.Context, params api.CreateJobParams) (ref.UUID, error)

	// UpdateJob updates the given job in the repository
	UpdateJob(ctx context.Context, j job.Job) (ref.UUID, error)

	// GetJob returns job with the given ID from the repository
	GetJob(ctx context.Context, ID ref.UUID) (job.Job, error)

	// ListJobs returns list of jobs from the repository
	ListJobs(ctx context.Context, paginationParams converters.PaginationParams) ([]job.Job, error)
}
