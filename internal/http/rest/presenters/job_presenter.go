package presenters

import (
	"net/http"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/api"
	"go.uber.org/zap"
)

// NewJobPresenter creates new job presentation service
func NewJobPresenter(logger *zap.SugaredLogger, serverAddr string) JobPresenter {
	return &jobPresenter{
		BasicPresenter: NewBasicPresenter(logger, serverAddr),
	}
}

type jobPresenter struct {
	*BasicPresenter
}

func (p jobPresenter) RenderJob(w http.ResponseWriter, job job.Job) {
	apiJob := p.convertJobToAPI(job)
	p.renderJSON(w, apiJob)
}

func (p jobPresenter) RenderJobList(w http.ResponseWriter, jobList []job.Job) {
	apiList := make([]api.Job, 0)

	for _, j := range jobList {
		apiJob := p.convertJobToAPI(j)
		apiList = append(apiList, apiJob)
	}

	p.renderJSON(w, apiList)
}

func (p jobPresenter) convertJobToAPI(j job.Job) api.Job {
	apiJob := api.Job{
		UUID:                       j.UUID().String(),
		CreatedAt:                  j.CreatedAt.String(),
		ProcessingStartedAt:        j.ProcessingStartedAt.String(),
		ChannelsDownloadStatus:     j.ChannelsDownloadStatus,
		ChannelsDownloadStartedAt:  j.ChannelsDownloadStartedAt.String(),
		ChannelsDownloadFinishedAt: j.ChannelsDownloadFinishedAt.String(),
		UsersDownloadStartedAt:     j.UsersDownloadStartedAt.String(),
		UsersDownloadFinishedAt:    j.UsersDownloadFinishedAt.String(),
		TicketsDownloadStartedAt:   j.TicketsDownloadStartedAt.String(),
		TicketsDownloadFinishedAt:  j.TicketsDownloadFinishedAt.String(),
		FinalStatus:                j.FinalStatus,
	}

	return apiJob
}
