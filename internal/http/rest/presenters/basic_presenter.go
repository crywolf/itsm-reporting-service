package presenters

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"go.uber.org/zap"
)

// NewBasicPresenter returns presentation service with basic functionality
func NewBasicPresenter(logger *zap.SugaredLogger, serverAddr string) *BasicPresenter {
	return &BasicPresenter{
		logger:     logger,
		serverAddr: serverAddr,
	}
}

// BasicPresenter must be included in all derived presenters via object composition
type BasicPresenter struct {
	logger     *zap.SugaredLogger
	serverAddr string
}

// RenderCreatedHeader sends Location header containing URI in the form 'route/resourceID'.
// Use it for rendering location of newly created resource
func (p BasicPresenter) RenderCreatedHeader(w http.ResponseWriter, route string, resourceID ref.UUID) {
	resourceURI := fmt.Sprintf("%s%s/%s", p.serverAddr, route, resourceID)

	w.Header().Set("Location", resourceURI)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// RenderNoContentHeader sends Location header containing URI in the form 'route/resourceID'.
// Use it for rendering location of updated resource
func (p BasicPresenter) RenderNoContentHeader(w http.ResponseWriter, route string, resourceID ref.UUID) {
	resourceURI := fmt.Sprintf("%s%s/%s", p.serverAddr, route, resourceID)

	w.Header().Set("Location", resourceURI)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// RenderError replies to the request with the specified error message and HTTP code
func (p BasicPresenter) RenderError(w http.ResponseWriter, msg string, err error) {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		if msg == "" {
			msg = httpErr.Error()
		}
		p.renderErrorJSON(w, msg, httpErr.Code())
		return
	}

	status := http.StatusInternalServerError

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		msg = fmt.Sprintf("internal error: %s", err.Error())
	} else {
		if msg == "" {
			msg = dErr.Error()
		}

		switch dErr.Code() {
		case domain.ErrorCodeInvalidArgument:
			status = http.StatusBadRequest
		case domain.ErrorCodeNotFound:
			status = http.StatusNotFound
		case domain.ErrorJobsProcessorIsBusy:
			status = http.StatusTooManyRequests
			w.Header().Set("Retry-After", "600")
		case domain.ErrorCodeUnknown:
			fallthrough
		default:
			status = http.StatusInternalServerError
		}
	}

	p.renderErrorJSON(w, msg, status)
}

// renderErrorJSON replies to the request with the specified error message and HTTP code.
// It encodes error string as JSON object {"error":"error_string"} and sets correct header.
// It does not otherwise end the request; the caller should ensure no further writes are done to 'w'.
// The error message should be plain text.
func (p BasicPresenter) renderErrorJSON(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	errorJSON, err := json.Marshal(msg)
	if err != nil {
		p.logger.Errorw("sending json error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(code)
	_, _ = fmt.Fprintf(w, `{"error":%s}`+"\n", errorJSON)
}

// renderJSON encodes 'v' to JSON and writes it to the 'w'. Also sets correct Content-Type header.
// It does not otherwise end the request; the caller should ensure no further writes are done to 'w'.
func (p BasicPresenter) renderJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		err = domain.WrapErrorf(err, domain.ErrorCodeUnknown, "could not encode JSON response")
		p.logger.Errorw("encoding json", "error", err)
		p.renderErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
