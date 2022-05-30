package converters

import (
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/api"
)

// PaginationParams provides information about current requested page number and a number of items per page to be displayed
type PaginationParams interface {
	// Page is the requested page number to be returned
	Page() uint

	// ItemsPerPage returns how many items per page should be displayed
	ItemsPerPage() uint
}

// JobPayloadConverter provides conversion from JSON request body payload to object
type JobPayloadConverter interface {
	// JobCreateParamsFromBody converts JSON payload to api.CreateJobParams
	JobCreateParamsFromBody(r *http.Request) (api.CreateJobParams, error)
}
