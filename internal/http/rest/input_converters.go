package rest

import (
	converters "github.com/KompiTech/itsm-reporting-service/internal/http/rest/api/input_converters"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/api/input_converters/validators"
)

func (s *Server) registerInputConverters() {
	validator := validators.NewPayloadValidator()

	s.jobInputPayloadConverter = converters.NewJobPayloadConverter(s.logger, validator)
}
