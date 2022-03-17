package job

import "fmt"

type Processor interface {
	WaitForJobs()
	ProcessNewJob()
}

func NewProcessor() Processor {
	return &processor{
		jobQueue: make(chan struct{}, 1),
	}
}

type processor struct {
	jobQueue chan struct{}
}

func (p *processor) WaitForJobs() {
	go func() {
		fmt.Println("WaitForJobs")
		for range p.jobQueue {
			fmt.Println("job was read from the queue - process it here!")
		}
	}()
}

func (p *processor) ProcessNewJob() {
	p.jobQueue <- struct{}{}
	fmt.Println("ProcessNewJob - job was inserted to the queue")
}
