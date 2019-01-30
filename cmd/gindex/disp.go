package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
)

// NewWorker creates takes a numeric id and a channel w/ worker pool.
func NewWorker(id int, workerPool chan chan IndexJob) Worker {
	return Worker{
		Id:         id,
		JobQueue:   make(chan IndexJob),
		WorkerPool: workerPool,
		QuitChan:   make(chan bool),
	}
}

type IndexJob struct {
	Rec   *httptest.ResponseRecorder
	Req   *http.Request
	Els   *ElServer
	Rpath *string
	Wg    *sync.WaitGroup
}

type Worker struct {
	Id         int
	JobQueue   chan IndexJob
	WorkerPool chan chan IndexJob
	QuitChan   chan bool
}

func (w *Worker) start() {
	go func() {
		for {
			// Add my jobQueue to the worker pool.
			w.WorkerPool <- w.JobQueue
			select {
			case job := <-w.JobQueue:
				// Dispatcher has added a job to my jobQueue.
				ReIndexRepo(job.Rec, job.Req, job.Els, job.Rpath)
				job.Wg.Done()
			case <-w.QuitChan:
				// We have been asked to stop.
				return
			}
		}
	}()
}

func (w *Worker) stop() {
	go func() {
		w.QuitChan <- true
	}()
}

// NewDispatcher creates, and returns a new Dispatcher object.
func NewDispatcher(jobQueue chan IndexJob, maxWorkers int) *Dispatcher {
	workerPool := make(chan chan IndexJob, maxWorkers)

	return &Dispatcher{
		jobQueue:   jobQueue,
		maxWorkers: maxWorkers,
		workerPool: workerPool,
	}
}

type Dispatcher struct {
	workerPool chan chan IndexJob
	maxWorkers int
	jobQueue   chan IndexJob
}

func (d *Dispatcher) Run(makeWorker func(int, chan chan IndexJob) Worker) {
	for i := 0; i < d.maxWorkers; i++ {
		worker := makeWorker(i+1, d.workerPool)
		worker.start()
	}
	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			go func() {
				workerJobQueue := <-d.workerPool
				workerJobQueue <- job
			}()
		}
	}
}
