package mocks

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/stretchr/testify/mock"
)

// JobServiceMock is a job service mock
type JobServiceMock struct {
	mock.Mock
}

func (s *JobServiceMock) CreateJob(_ context.Context) (ref.UUID, error) {
	args := s.Called()
	return args.Get(0).(ref.UUID), args.Error(1)
}

func (s *JobServiceMock) GetJob(_ context.Context, ID ref.UUID) (job.Job, error) {
	args := s.Called(ID)
	return args.Get(0).(job.Job), args.Error(1)
}

func (s *JobServiceMock) ListJobs(_ context.Context) ([]job.Job, error) {
	args := s.Called()
	return args.Get(0).([]job.Job), args.Error(1)
}