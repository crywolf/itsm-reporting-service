package rest

import (
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/presenters"
)

func (s *Server) registerRoutes() {
	s.router.POST("/jobs", s.CreateJob())
	s.router.GET("/jobs/:id", s.GetJob())
	s.router.GET("/jobs", s.ListJobs())

	// default Not Found handler
	s.router.NotFound = http.HandlerFunc(s.JSONNotFoundError)
}

// JSONNotFoundError replies to the request with the 404 page not found general error message
// in JSON format and sets correct header and HTTP code
func (s Server) JSONNotFoundError(w http.ResponseWriter, _ *http.Request) {
	s.jobsPresenter.RenderError(w, "", presenters.NewErrorf(http.StatusNotFound, "404 page not found"))
}
