package memory

// Job stored in memory storage
type Job struct {
	ID string

	CreatedAt string

	ProcessingStartedAt string

	ChannelsDownloadStatus string
}
