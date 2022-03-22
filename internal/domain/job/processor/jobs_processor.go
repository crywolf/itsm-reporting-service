package jobprocessor

import (
	"context"
	"sync"
	"time"

	"github.com/KompiTech/itsm-reporting-service/internal/domain"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
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
func NewJobProcessor(logger *zap.SugaredLogger, jobRepository repository.JobRepository) JobProcessor {
	return &processor{
		logger:   logger,
		repo:     jobRepository,
		jobQueue: make(chan struct{}, 1),
	}
}

type processor struct {
	logger   *zap.SugaredLogger
	repo     repository.JobRepository
	jobQueue chan struct{}
	mu       sync.Mutex
	ready    bool
}

// WaitForJobs starts job queue loop and when new job appears starts processing it
func (p *processor) WaitForJobs() {
	c := make(chan struct{}, 1)

	go func() {
		p.mu.Lock()
		p.ready = true
		p.mu.Unlock()

		c <- struct{}{}

		for range p.jobQueue {
			// new job was inserted to the repository => process it
			ctx := context.Background()
			j, err := p.repo.GetLastJob(ctx)
			if err != nil {
				p.logger.Errorw("Getting last job from the queue failed", "error", err)
			}
			p.logger.Infow("New job read from the queue", "time", time.Now().Format(time.RFC3339), "id", j.UUID())
			p.downloadChannelList(j.UUID())
			p.downloadUsersFromChannels(j.UUID())
			p.downloadTicketsFromChannels(j.UUID())

			p.mu.Lock()
			p.ready = true
			p.mu.Unlock()
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

func (p *processor) downloadChannelList(jobID ref.UUID) {
	p.logger.Infow("Starting channel list download", "time", time.Now().Format(time.RFC3339), "job", jobID)
	//	time.Sleep(1 * time.Second)
	p.logger.Infow("Channel list downloaded", "time", time.Now().Format(time.RFC3339), "job", jobID)
}

func (p *processor) downloadUsersFromChannels(jobID ref.UUID) {
	p.logger.Infow("Starting users download", "time", time.Now().Format(time.RFC3339), "job", jobID)
	//	time.Sleep(1 * time.Second)
	p.logger.Infow("Users downloaded", "time", time.Now().Format(time.RFC3339), "job", jobID)
}

func (p *processor) downloadTicketsFromChannels(jobID ref.UUID) {
	p.logger.Infow("Starting tickets download", "time", time.Now().Format(time.RFC3339), "job", jobID)
	// 	time.Sleep(10 * time.Second)
	p.logger.Infow("Tickets downloaded", "time", time.Now().Format(time.RFC3339), "job", jobID)
}
