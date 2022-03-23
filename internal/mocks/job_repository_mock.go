package mocks

import (
	"context"
	"sync"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/stretchr/testify/mock"
)

// JobRepositoryMock is a job repository mock
type JobRepositoryMock struct {
	mock.Mock
	Wg sync.WaitGroup
}

func (m *JobRepositoryMock) AddJob(_ context.Context, job job.Job) (ref.UUID, error) {
	args := m.Called(job)
	return args.Get(0).(ref.UUID), args.Error(1)
}

func (m *JobRepositoryMock) UpdateJob(_ context.Context, job job.Job) (ref.UUID, error) {
	//TODO implement me
	panic("implement me")
}

func (m *JobRepositoryMock) GetJob(_ context.Context, ID ref.UUID) (job.Job, error) {
	//TODO implement me
	panic("implement me")
}

func (m *JobRepositoryMock) GetLastJob(_ context.Context) (job.Job, error) {
	defer m.Wg.Done()
	args := m.Called()
	return args.Get(0).(job.Job), args.Error(1)
}

func (m *JobRepositoryMock) ListJobs(_ context.Context) ([]job.Job, error) {
	//TODO implement me
	panic("implement me")
}
