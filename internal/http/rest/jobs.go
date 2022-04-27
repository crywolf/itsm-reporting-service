package rest

import (
	"fmt"
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/presenters"
	"github.com/julienschmidt/httprouter"
)

// swagger:route POST /jobs jobs CreateJob
// Creates a new job
// responses:
//	201: jobCreatedResponse
//	429: errorResponse429

// CreateJob returns handler for creating new job
func (s *Server) CreateJob() func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()
		newID, err := s.jobsService.CreateJob(ctx)
		if err != nil {
			s.logger.Errorw("CreateJob handler failed", "error", err)
			s.jobsPresenter.RenderError(w, "", err)
			return
		}

		if jobErr := s.jobsProcessor.ProcessNewJob(newID); jobErr != nil {
			s.logger.Errorw("CreateJob handler failed", "error", jobErr)

			// mark the unprocessed job as failed
			j, err := s.jobsService.GetJob(ctx, newID)
			if err != nil {
				s.logger.Errorw("CreateJob handler failed", "error", err)
			}

			j.FinalStatus = fmt.Sprintf("Error: %s", jobErr)

			_, err = s.jobsService.UpdateJob(ctx, j)
			if err != nil {
				s.logger.Errorw("CreateJob handler failed", "error", err)
			}

			s.jobsPresenter.RenderError(w, "", jobErr)
			return
		}

		s.jobsPresenter.RenderCreatedHeader(w, listJobsRoute, newID)
	}
}

const listJobsRoute = "/jobs"

// swagger:route GET /jobs jobs ListJobs
// Returns a list of jobs
// responses:
//	200: jobListResponse

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

// swagger:route GET /jobs/{uuid} jobs GetJob
// Returns a single job from the repository
// responses:
//	200: jobResponse
//	404: errorResponse404

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
			s.logger.Errorw("GetJob handler failed", "ID", id, "error", err)
			s.jobsPresenter.RenderError(w, "job not found", err)
			return
		}

		s.jobsPresenter.RenderJob(w, j)
	}
}
