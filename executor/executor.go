package executor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Rx-11/go-wasp/internal"
	"github.com/Rx-11/go-wasp/registry"
)

var (
	ErrQueueFull    = errors.New("queue full")
	ErrShuttingDown = errors.New("executor shutting down")
)

type Job struct {
	Input      map[string]any
	ResultChan chan Result
}

type Result struct {
	Out map[string]any
	Err error
}

type FunctionExecutor struct {
	name            string
	reg             registry.Registry
	queue           chan *Job
	workers         int
	stopChan        chan struct{}
	globalSemaphore chan struct{}
}

func NewFunctionExecutor(name string, reg registry.Registry, queueSize, workers int, globalSemaphore chan struct{}) *FunctionExecutor {
	if workers <= 0 {
		workers = 1
	}
	ex := &FunctionExecutor{
		name:            name,
		reg:             reg,
		queue:           make(chan *Job, queueSize),
		workers:         workers,
		stopChan:        make(chan struct{}),
		globalSemaphore: globalSemaphore,
	}
	for i := 0; i < workers; i++ {
		go ex.worker()
	}
	return ex
}

func (e *FunctionExecutor) Enqueue(ctx context.Context, input map[string]any) (map[string]any, error) {
	job := &Job{
		Input:      input,
		ResultChan: make(chan Result, 1),
	}
	time.Sleep(2 * time.Millisecond)
	select {
	case e.queue <- job:
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, ErrQueueFull
	}

	select {
	case res := <-job.ResultChan:
		return res.Out, res.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (e *FunctionExecutor) worker() {
	for {
		select {
		case <-e.stopChan:
			return
		case job := <-e.queue:
			e.runJob(job)
		}
	}
}

func (e *FunctionExecutor) runJob(job *Job) {
	if e.globalSemaphore != nil {
		select {
		case e.globalSemaphore <- struct{}{}:
			defer func() { <-e.globalSemaphore }()
		default:
			e.globalSemaphore <- struct{}{}
			defer func() { <-e.globalSemaphore }()
		}
	}

	wasmBytes, err := e.reg.GetFunction(e.name)
	if err != nil {
		job.ResultChan <- Result{Err: fmt.Errorf("get wasm: %w", err)}
		return
	}

	out, err := internal.ExecuteWASM(wasmBytes, job.Input)
	job.ResultChan <- Result{Out: out, Err: err}
}

func (e *FunctionExecutor) Stop() {
	close(e.stopChan)
}
