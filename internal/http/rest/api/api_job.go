package api

// Job API object
// swagger:model
type Job struct {
	// required: true
	// swagger:strfmt uuid
	UUID string `json:"uuid"`

	// Time when the job was created
	// required: true
	// swagger:strfmt date-time
	CreatedAt string `json:"created_at,omitempty"`

	// Time when the job processing started
	// swagger:strfmt date-time
	ProcessingStartedAt string `json:"processing_started_at,omitempty"`

	// Status of the channel list download (success/error)
	ChannelsDownloadStatus string `json:"channels_download_status,omitempty"`
}
