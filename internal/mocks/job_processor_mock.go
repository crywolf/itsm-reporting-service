package mocks

import (
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/stretchr/testify/mock"
)

// JobProcessorMock is a job processor mock
type JobProcessorMock struct {
	mock.Mock
}

func (p *JobProcessorMock) WaitForJobs() {}

func (p *JobProcessorMock) ProcessNewJob(jobID ref.UUID) error {
	args := p.Called(jobID)
	return args.Error(0)
}
