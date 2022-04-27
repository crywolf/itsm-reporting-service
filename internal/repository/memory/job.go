package memory

// Job stored in memory storage
type Job struct {
	ID string

	CreatedAt string

	ChannelsDownloadStartedAt string

	ChannelsDownloadFinishedAt string

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
