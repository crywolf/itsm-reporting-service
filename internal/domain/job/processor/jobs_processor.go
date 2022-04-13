package jobprocessor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	chandownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/channel/downloader"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/excel"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	ticketdownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/ticket/downloader"
	userdownloader "github.com/KompiTech/itsm-reporting-service/internal/domain/user/downloader"
	"github.com/KompiTech/itsm-reporting-service/internal/repository"
	"go.uber.org/zap"
)

// JobProcessor processes created jobs
type JobProcessor interface {
	WaitForJobs()
	ProcessNewJob(jobID ref.UUID) error
}

// ErrorBusy is returned when job processor is busy doing its job
var ErrorBusy = domain.NewErrorf(domain.ErrorJobsProcessorIsBusy, "job is being processed, try it later")

// NewJobProcessor returns job processor. Manually call WaitForJobs() to start it.
func NewJobProcessor(
	logger *zap.SugaredLogger,
	jobRepository repository.JobRepository,
	channelDownloader chandownloader.ChannelDownloader,
	userDownloader userdownloader.UserDownloader,
	ticketDownloader ticketdownloader.TicketDownloader,
	excelGenerator excel.Generator,
) JobProcessor {
	return &processor{
		logger:            logger,
		jobRepository:     jobRepository,
		channelDownloader: channelDownloader,
		userDownloader:    userDownloader,
		ticketDownloader:  ticketDownloader,
		excelGenerator:    excelGenerator,
		jobQueue:          make(chan struct{}, 1),
	}
}

type processor struct {
	logger            *zap.SugaredLogger
	jobRepository     repository.JobRepository
	channelDownloader chandownloader.ChannelDownloader
	userDownloader    userdownloader.UserDownloader
	ticketDownloader  ticketdownloader.TicketDownloader
	excelGenerator    excel.Generator
	jobQueue          chan struct{}
	mu                sync.Mutex
	ready             bool
}

// WaitForJobs starts job queue loop and when new job appears starts processing it
func (p *processor) WaitForJobs() {
	c := make(chan struct{}, 1)

	go func() {
		p.markAsReady()

		c <- struct{}{}

		for range p.jobQueue {
			// new job was inserted to the repository => process it
			ctx := context.Background()
			j, err := p.jobRepository.GetLastJob(ctx)
			if err != nil {
				p.logger.Errorw("Getting last job from the queue failed", "error", err)
				p.markJobAsFailed(ctx, j.UUID(), err)
				p.markAsReady()
				continue
			}
			p.logger.Infow("New job read from the queue", "time", time.Now().Format(time.RFC3339), "id", j.UUID())

			if err := p.userDownloader.Reset(ctx); err != nil {
				p.logger.Errorw("User downloader reset failed", "error", err)
				p.markJobAsFailed(ctx, j.UUID(), err)
				p.markAsReady()
				continue
			}

			if err := p.ticketDownloader.Reset(ctx); err != nil {
				p.logger.Errorw("Ticket downloader reset failed", "error", err)
				p.markJobAsFailed(ctx, j.UUID(), err)
				p.markAsReady()
				continue
			}

			if err := p.downloadChannelList(ctx, j.UUID()); err != nil {
				p.logger.Errorw("Channels download failed", "error", err)
				p.markJobAsFailed(ctx, j.UUID(), err)
				p.markAsReady()
				continue
			}

			if err := p.downloadUsersFromChannels(ctx, j.UUID()); err != nil {
				p.logger.Errorw("Users download failed", "error", err)
				p.markJobAsFailed(ctx, j.UUID(), err)
				p.markAsReady()
				continue
			}

			if err := p.downloadTicketsFromChannels(ctx, j.UUID()); err != nil {
				p.logger.Errorw("Tickets download failed", "error", err)
				p.markJobAsFailed(ctx, j.UUID(), err)
				p.markAsReady()
				continue
			}

			if err := p.generateExcelFiles(ctx, j.UUID()); err != nil {
				p.logger.Errorw("Excel files generation failed", "error", err)
				p.markJobAsFailed(ctx, j.UUID(), err)
				p.markAsReady()
				continue
			}

			// TODO Send emails

			p.markJobAsFinished(ctx, j.UUID())
			p.markAsReady()
		}
	}()

	<-c // wait for goroutine to start
	p.logger.Info("Jobs processor is waiting for new jobs")
}

// ProcessNewJob notifies the job processor about new job
func (p *processor) ProcessNewJob(jobID ref.UUID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.ready {
		return ErrorBusy
	}

	p.jobQueue <- struct{}{}

	p.ready = false
	p.logger.Infow("New job inserted to the queue", "time", time.Now().Format(time.RFC3339), "job", jobID)
	return nil
}

func (p *processor) downloadChannelList(ctx context.Context, jobID ref.UUID) error {
	p.logger.Infow("Starting channel list download", "time", time.Now().Format(time.RFC3339), "job", jobID)

	j, err := p.jobRepository.GetJob(ctx, jobID)
	if err != nil {
		p.logger.Errorw("Could not get job for channel download update", "error", err)
	}

	j.ChannelsDownloadStartedAt.SetNow()

	if _, err := p.jobRepository.UpdateJob(ctx, j); err != nil {
		p.logger.Errorw("Could not mark job as channel download started", "error", err)
	}

	if err := p.channelDownloader.DownloadChannelList(ctx); err != nil {
		return err
	}

	j.ChannelsDownloadFinishedAt.SetNow()

	if _, err := p.jobRepository.UpdateJob(ctx, j); err != nil {
		p.logger.Errorw("Could not mark job as channel download finished", "error", err)
	}

	p.logger.Infow("Channel list downloaded", "time", time.Now().Format(time.RFC3339), "job", jobID)
	return nil
}

func (p *processor) downloadUsersFromChannels(ctx context.Context, jobID ref.UUID) error {
	p.logger.Infow("Starting users download", "time", time.Now().Format(time.RFC3339), "job", jobID)

	j, err := p.jobRepository.GetJob(ctx, jobID)
	if err != nil {
		p.logger.Errorw("Could not get job for user download update", "error", err)
	}

	j.UsersDownloadStartedAt.SetNow()

	if _, err := p.jobRepository.UpdateJob(ctx, j); err != nil {
		p.logger.Errorw("Could not mark job as user download started", "error", err)
	}

	if err := p.userDownloader.DownloadUsers(ctx); err != nil {
		return err
	}

	j.UsersDownloadFinishedAt.SetNow()

	if _, err := p.jobRepository.UpdateJob(ctx, j); err != nil {
		p.logger.Errorw("Could not mark job as user download finished", "error", err)
	}

	p.logger.Infow("Users downloaded", "time", time.Now().Format(time.RFC3339), "job", jobID)
	return nil
}

func (p *processor) downloadTicketsFromChannels(ctx context.Context, jobID ref.UUID) error {
	p.logger.Infow("Starting tickets download", "time", time.Now().Format(time.RFC3339), "job", jobID)

	j, err := p.jobRepository.GetJob(ctx, jobID)
	if err != nil {
		p.logger.Errorw("Could not get job for ticket download update", "error", err)
	}

	j.TicketsDownloadStartedAt.SetNow()

	if _, err := p.jobRepository.UpdateJob(ctx, j); err != nil {
		p.logger.Errorw("Could not mark job as ticket download started", "error", err)
	}

	if err := p.ticketDownloader.DownloadTickets(ctx); err != nil {
		return err
	}

	j.TicketsDownloadFinishedAt.SetNow()

	if _, err := p.jobRepository.UpdateJob(ctx, j); err != nil {
		p.logger.Errorw("Could not mark job as ticket download finished", "error", err)
	}

	p.logger.Infow("Tickets downloaded", "time", time.Now().Format(time.RFC3339), "job", jobID)
	return nil
}

func (p *processor) generateExcelFiles(ctx context.Context, jobID ref.UUID) error {
	p.logger.Infow("Starting Excel files generation", "time", time.Now().Format(time.RFC3339), "job", jobID)

	j, err := p.jobRepository.GetJob(ctx, jobID)
	if err != nil {
		p.logger.Errorw("Could not get job for Excel files generation update", "error", err)
	}

	j.ExcelFilesGenerationStartedAt.SetNow()

	if _, err := p.jobRepository.UpdateJob(ctx, j); err != nil {
		p.logger.Errorw("Could not mark job as Excel files generation started", "error", err)
	}

	if err := p.excelGenerator.GenerateExcelFiles(ctx); err != nil {
		return err
	}

	j.ExcelFilesGenerationFinishedAt.SetNow()

	if _, err := p.jobRepository.UpdateJob(ctx, j); err != nil {
		p.logger.Errorw("Could not mark job as Excel files generation finished", "error", err)
	}

	p.logger.Infow("Excel files generated", "time", time.Now().Format(time.RFC3339), "job", jobID)
	return nil
}

func (p *processor) markAsReady() {
	p.mu.Lock()
	p.ready = true
	p.mu.Unlock()
}

func (p *processor) markJobAsFailed(ctx context.Context, jobID ref.UUID, jobErr error) {
	j, err := p.jobRepository.GetJob(ctx, jobID)
	if err != nil {
		p.logger.Errorw("Could not mark job as failed", "error", err)
	}

	j.FinalStatus = fmt.Sprintf("Error: %s", jobErr)

	if _, err := p.jobRepository.UpdateJob(ctx, j); err != nil {
		p.logger.Errorw("Could not mark job as failed", "error", err)
	}
}

func (p *processor) markJobAsFinished(ctx context.Context, jobID ref.UUID) {
	j, err := p.jobRepository.GetJob(ctx, jobID)
	if err != nil {
		p.logger.Errorw("Could not mark job as finished", "error", err)
	}

	j.FinalStatus = "Success"

	if _, err := p.jobRepository.UpdateJob(ctx, j); err != nil {
		p.logger.Errorw("Could not mark job as finished", "error", err)
	}
}
