package rest

import "github.com/KompiTech/itsm-reporting-service/internal/http/rest/presenters"

func (s *Server) registerPresenters() {
	s.jobsPresenter = presenters.NewJobPresenter(s.logger, s.ExternalLocationAddress)
}
