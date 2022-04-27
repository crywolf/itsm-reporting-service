// Package api ITSM Reporting REST API
//
// Documentation for ITSM Reporting Service REST API
//
//	Schemes: http
//	BasePath: /
//	Version: 0.0.1
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
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
	ProcessingStartedAt string `json:"processing_started_at,omitempty"` // TODO remove - unnecessary

	// Status of the channel list download (success/error)
	ChannelsDownloadStatus string `json:"channels_download_status,omitempty"` // TODO remove - unnecessary

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

	// Time when Excel files generation started
	// swagger:strfmt date-time
	ExcelFilesGenerationStartedAt string `json:"excel_files_generation_started_at,omitempty"`

	// Time when Excel files generation finished
	// swagger:strfmt date-time
	ExcelFilesGenerationFinishedAt string `json:"excel_files_generation_finished_at,omitempty"`

	// Time when sending of emails started
	// swagger:strfmt date-time
	EmailsSendingStartedAt string `json:"emails_sending_started_at,omitempty"`

	// Time when sending of emails finished
	// swagger:strfmt date-time
	EmailsSendingFinishedAt string `json:"emails_sending_finished_at,omitempty"`

	// Status of the finished job (success/error)
	FinalStatus string `json:"final_status,omitempty"`
}

// NOTE: Types defined here are purely for documentation purposes
// these types are not used by any of the handlers

/*// UUID represents UUID of a resource
// swagger:strfmt uuid
type UUID string
*/

// Data structure representing a single job
// swagger:response jobResponse
type jobResponseWrapper struct {
	// in: body
	Body Job
}

// A list of jobs
// swagger:response jobListResponse
type jobListResponseWrapper struct {
	// in: body
	Body []Job
}

// Created
// swagger:response jobCreatedResponse
type jobCreatedResponseWrapper struct {
	// URI of the resource
	// example: http://localhost:8080/jobs/2af4f493-0bd5-4513-b440-6cbb465feadb
	// in: header
	Location string
}

// Error
// swagger:response errorResponse
type errorResponseWrapper struct {
	// in: body
	Body struct {
		// required: true
		// Description of the error
		ErrorMessage string `json:"error"`
	}
}

// Not Found
// swagger:response errorResponse404
type errorResponseWrapper404 errorResponseWrapper

// Error Too Many Requests
// swagger:response error429Response
type errorResponse429Wrapper struct {
	// example: Retry-After: 600
	// in: header
	RetryAfter string `json:"Retry-After"`

	// in: body
	Body struct {
		// required: true
		// Description of the error
		ErrorMessage string `json:"error"`
	}
}

// Too Many Requests
// swagger:response errorResponse429
type errorResponseWrapper429 errorResponse429Wrapper
