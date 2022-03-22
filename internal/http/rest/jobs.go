package rest

import (
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/presenters"
	"github.com/julienschmidt/httprouter"
)

// CreateJob returns handler for creating new job
func (s *Server) CreateJob() func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		newID, err := s.jobsService.CreateJob(r.Context())
		if err != nil {
			s.logger.Errorw("CreateJob handler failed", "error", err)
			s.jobsPresenter.RenderError(w, "", err)
			return
		}

		if err = s.jobsProcessor.ProcessNewJob(newID); err != nil {
			s.logger.Errorw("CreateJob handler failed", "error", err)
			s.jobsPresenter.RenderError(w, "", err)
			return
		}

		s.jobsPresenter.RenderCreatedHeader(w, listJobsRoute, newID)
	}
}

const listJobsRoute = "/jobs"

// ListJobs returns handler for listing jobs
func (s *Server) ListJobs() func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		list, err := s.jobsService.ListJobs(r.Context())
		if err != nil {
			s.logger.Errorw("ListJobs handler failed", "error", err)
			s.jobsPresenter.RenderError(w, "", err)
			return
		}

		s.jobsPresenter.RenderJobList(w, list)
	}
}

// GetJob returns handler for getting single job
func (s *Server) GetJob() func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		if id == "" {
			err := presenters.NewErrorf(http.StatusBadRequest, "malformed URL: missing resource ID param")
			s.logger.Errorw("GetJob handler failed", "error", err)
			s.jobsPresenter.RenderError(w, "", err)
			return
		}

		j, err := s.jobsService.GetJob(r.Context(), ref.UUID(id))
		if err != nil {
			s.logger.Errorw("GetIncident handler failed", "ID", id, "error", err)
			s.jobsPresenter.RenderError(w, "job not found", err)
			return
		}

		s.jobsPresenter.RenderJob(w, j)
	}
}
