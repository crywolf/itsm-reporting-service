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

func (r *JobRepositoryMock) AddJob(_ context.Context, job job.Job) (ref.UUID, error) {
	args := r.Called(job)
	return args.Get(0).(ref.UUID), args.Error(1)
}

func (r *JobRepositoryMock) UpdateJob(ctx context.Context, job job.Job) (ref.UUID, error) {
	//TODO implement me
	panic("implement me")
}

func (r *JobRepositoryMock) GetJob(ctx context.Context, ID ref.UUID) (job.Job, error) {
	//TODO implement me
	panic("implement me")
}

func (r *JobRepositoryMock) GetLastJob(_ context.Context) (job.Job, error) {
	defer r.Wg.Done()
	args := r.Called()
	return args.Get(0).(job.Job), args.Error(1)
}

func (r *JobRepositoryMock) ListJobs(ctx context.Context) ([]job.Job, error) {
	//TODO implement me
	panic("implement me")
}
