package mocks

// JobProcessorMock is a job processor mock
type JobProcessorMock struct{}

func (p *JobProcessorMock) WaitForJobs() {}

func (p *JobProcessorMock) ProcessNewJob() {}
