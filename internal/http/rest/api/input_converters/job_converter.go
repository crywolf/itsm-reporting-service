package converters

import (
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/api"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/api/input_converters/validators"
	"go.uber.org/zap"
)

// NewJobPayloadConverter creates an job input payload converting service
func NewJobPayloadConverter(logger *zap.SugaredLogger, validator validators.PayloadValidator) JobPayloadConverter {
	return &jobPayloadConverter{
		BasePayloadConverter: NewBasePayloadConverter(logger, validator),
	}
}

type jobPayloadConverter struct {
	*BasePayloadConverter
}

// JobCreateParamsFromBody converts JSON payload to api.CreateJobParams
func (c jobPayloadConverter) JobCreateParamsFromBody(r *http.Request) (api.CreateJobParams, error) {
	var payload api.CreateJobParams

	if err := c.unmarshalFromBody(r, &payload); err != nil {
		return payload, err
	}

	return payload, nil
}
