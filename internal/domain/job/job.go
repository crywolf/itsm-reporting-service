package job

import (
	"fmt"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/types"
)

// Job domain object
type Job struct {
	uuid ref.UUID

	// Time when the job was created
	CreatedAt types.DateTime

	// Time when the job processing started
	ProcessingStartedAt types.DateTime

	// Time when the channels download started
	ChannelsDownloadStartedAt types.DateTime

	// Time when the channels download finished
	ChannelsDownloadFinishedAt types.DateTime

	// Status of the channel list download (success/error)
	ChannelsDownloadStatus string // TODO remove - unnecessary

	// Time when the users download started
	UsersDownloadStartedAt types.DateTime

	// Time when the users download finished
	UsersDownloadFinishedAt types.DateTime

	// Time when the tickets download started
	TicketsDownloadStartedAt types.DateTime

	// Time when the tickets download finished
	TicketsDownloadFinishedAt types.DateTime

	// Time when Excel files generation started
	ExcelFilesGenerationStartedAt types.DateTime

	// Time when Excel files generation finished
	ExcelFilesGenerationFinishedAt types.DateTime

	// Status of the finished job (success/error)
	FinalStatus string
}

// UUID getter
func (e Job) UUID() ref.UUID {
	return e.uuid
}

// SetUUID returns error if UUID was already set
func (e *Job) SetUUID(v ref.UUID) error {
	if !e.uuid.IsZero() {
		return fmt.Errorf("job: cannot set UUID, it was already set (%s)", e.uuid)
	}
	e.uuid = v
	return nil
}
