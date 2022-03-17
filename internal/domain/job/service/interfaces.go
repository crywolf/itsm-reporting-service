package jobsvc

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
)

// JobService provides job operations
type JobService interface {
	// CreateJob creates new job and adds it to the repository
	CreateJob(ctx context.Context) (ref.UUID, error)

	// GetJob returns the job with the given ID from the repository
	GetJob(ctx context.Context, ID ref.UUID) (job.Job, error)

	// ListJobs returns the list of jobs from the repository
	ListJobs(ctx context.Context) ([]job.Job, error)
}
