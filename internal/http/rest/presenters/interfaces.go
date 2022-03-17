package presenters

import (
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
)

// ErrorPresenter allows replying with error
type ErrorPresenter interface {
	// RenderError replies to the request with the specified error message and HTTP code.
	// It does not otherwise end the request; the caller should ensure no further writes are done to 'w'.
	// The error message should be plain text.
	RenderError(w http.ResponseWriter, msg string, err error)
}

// LocationHeaderPresenter allows sending Location header with URI of the resource
type LocationHeaderPresenter interface {
	// RenderCreatedHeader sends Location header containing URI in the form 'route/resourceID'.
	// Use it for rendering location of newly created resource
	RenderCreatedHeader(w http.ResponseWriter, route string, resourceID ref.UUID)

	// RenderNoContentHeader sends Location header containing URI in the form 'route/resourceID'.
	// Use it for rendering location of updated resource
	RenderNoContentHeader(w http.ResponseWriter, route string, resourceID ref.UUID)
}

// JobPresenter provides REST responses for job resource
type JobPresenter interface {
	ErrorPresenter
	LocationHeaderPresenter

	// RenderJob encodes job and writes it to 'w'.  Also sets correct Content-Type header.
	// It does not otherwise end the request; the caller should ensure no further writes are done to 'w'.
	RenderJob(w http.ResponseWriter, job job.Job)

	// RenderJobList encodes list of jobs and writes it to 'w'.  Also sets correct Content-Type header.
	// It does not otherwise end the request; the caller should ensure no further writes are done to 'w'.
	RenderJobList(w http.ResponseWriter, jobList []job.Job)
}
