package repository

import (
	"context"
	"time"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/channel"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ticket"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/types"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/user"
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
	// StoreChannelList stores list of channels to the repository (rewrites the content of the repository)
	StoreChannelList(ctx context.Context, channelList channel.List) error

	// GetChannelList returns the list of channels from the repository
	GetChannelList(ctx context.Context) (channel.List, error)
}

// UserRepository provides access to the user repository
type UserRepository interface {
	// AddUserList adds list of users to the repository
	AddUserList(ctx context.Context, userList user.List) error

	// GetUserInChannel returns user from specified channel from the repository
	GetUserInChannel(ctx context.Context, channelID, userID string) (user.User, error)

	// Truncate removes all items from repository
	Truncate(ctx context.Context) error
}

// TicketRepository provides access to the ticket repository
type TicketRepository interface {
	// AddTicketList adds list of tickets to the repository
	AddTicketList(ctx context.Context, ticketList ticket.List) error

	// GetTicketsByEmailAddress returns tickets for the specified user's email address from the repository.
	// It sorts the returned list, first are Incidents, then Requests.
	GetTicketsByEmailAddress(ctx context.Context, userEmail string) (ticket.List, error)

	// GetTicketsByChannelID returns tickets for the specified channel from the repository.
	// It groups the returned list by user email address and sorts it, first are Incidents, then Requests.
	GetTicketsByChannelID(ctx context.Context, channelID string) (ticket.List, error)

	// GetDistinctEmailAddresses returns distinct email addresses from the repository
	GetDistinctEmailAddresses(ctx context.Context) ([]string, error)

	// GetDistinctChannelIDs returns distinct channel IDs from the repository
	GetDistinctChannelIDs(ctx context.Context) ([]string, error)

	// Truncate removes all items from the repository
	Truncate(ctx context.Context) error
}
