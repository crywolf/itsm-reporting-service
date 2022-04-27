package memory

// Job stored in memory storage
type Job struct {
	ID string

	CreatedAt string

	ProcessingStartedAt string // TODO remove - unnecessary

	ChannelsDownloadStartedAt string

	ChannelsDownloadFinishedAt string

	ChannelsDownloadStatus string // TODO remove - unnecessary

	UsersDownloadStartedAt string

	UsersDownloadFinishedAt string

	TicketsDownloadStartedAt string

	TicketsDownloadFinishedAt string

	ExcelFilesGenerationStartedAt string

	ExcelFilesGenerationFinishedAt string

	EmailsSendingStartedAt string

	EmailsSendingFinishedAt string

	FinalStatus string
}
