package repository

import (
	"context"
	"time"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/types"
)

// Clock provides Now method to enable mocking
type Clock interface {
	// Now returns current time
	Now() time.Time

	// NowFormatted returns time in RFC3339 format
	NowFormatted() types.DateTime
}

// JobRepository provides access to the jobs repository
type JobRepository interface {
	// AddJob adds the given job to the repository
	AddJob(ctx context.Context, job job.Job) (ref.UUID, error)

	// UpdateJob updates the given job in the repository
	UpdateJob(ctx context.Context, job job.Job) (ref.UUID, error)

	// GetJob returns the job with the given ID from the repository
	GetJob(ctx context.Context, ID ref.UUID) (job.Job, error)

	// GetLastJob returns the last inserted job from the repository
	GetLastJob(ctx context.Context) (job.Job, error)

	// ListJobs returns the list of jobs from the repository
	ListJobs(ctx context.Context) ([]job.Job, error)
}

// ChannelRepository provides access to the channel repository
type ChannelRepository interface {
	// StoreChannelList stores list of channels to the repository
	StoreChannelList(ctx context.Context, channelList channel.List) error

	// GetChannelList returns the list of channels from the repository
	GetChannelList(ctx context.Context) (channel.List, error)
}
