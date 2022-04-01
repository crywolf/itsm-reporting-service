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

	// Time when the channels download started
	// swagger:strfmt date-time
	ChannelsDownloadStartedAt string `json:"channels_download_started_at,omitempty"`

	// Time when the channels download finished
	// swagger:strfmt date-time
	ChannelsDownloadFinishedAt string `json:"channels_download_finished_at,omitempty"`

	// Time when the users download started
	// swagger:strfmt date-time
	UsersDownloadStartedAt string `json:"users_download_started_at,omitempty"`

	// Time when the users download finished
	// swagger:strfmt date-time
	UsersDownloadFinishedAt string `json:"users_download_finished_at,omitempty"`

	// Time when the tickets download started
	// swagger:strfmt date-time
	TicketsDownloadStartedAt string `json:"tickets_download_started_at,omitempty"`

	// Time when the tickets download finished
	// swagger:strfmt date-time
	TicketsDownloadFinishedAt string `json:"tickets_download_finished_at,omitempty"`

	// Status of the finished job (success/error)
	FinalStatus string `json:"final_status,omitempty"`
}
