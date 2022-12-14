package mocks

import (
	"context"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/stretchr/testify/mock"
)

// JobRepositoryMock is a job repository mock
type JobRepositoryMock struct {
	mock.Mock
}

func (m *JobRepositoryMock) AddJob(_ context.Context, job job.Job) (ref.UUID, error) {
	args := m.Called(job)
	return args.Get(0).(ref.UUID), args.Error(1)
}

func (m *JobRepositoryMock) UpdateJob(_ context.Context, job job.Job) (ref.UUID, error) {
	args := m.Called(job)
	return args.Get(0).(ref.UUID), args.Error(1)
}

func (m *JobRepositoryMock) GetJob(_ context.Context, ID ref.UUID) (job.Job, error) {
	args := m.Called(ID)
	return args.Get(0).(job.Job), args.Error(1)
}

func (m *JobRepositoryMock) GetLastJob(_ context.Context) (job.Job, error) {
	args := m.Called()
	return args.Get(0).(job.Job), args.Error(1)
}

func (m *JobRepositoryMock) ListJobs(_ context.Context, _, _ uint) ([]job.Job, error) {
	//TODO implement me
	panic("implement me")
}
