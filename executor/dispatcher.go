package executor

import (
	"context"
	"sync"
	"time"

	"github.com/Rx-11/go-wasp/registry"
)

type Dispatcher struct {
	globalLimiter chan struct{}
	executors     map[string]*FunctionExecutor
	mu            sync.RWMutex

	queueSize int
	workers   int
	timeout   time.Duration

	reg registry.Registry
}

func NewDispatcher(reg registry.Registry, queueSize, workers int, nodeMax int, invokeTimeout time.Duration) *Dispatcher {
	var limit chan struct{}
	if nodeMax > 0 {
		limit = make(chan struct{}, nodeMax)
	}

	return &Dispatcher{
		globalLimiter: limit,
		executors:     make(map[string]*FunctionExecutor),
		queueSize:     queueSize,
		workers:       workers,
		timeout:       invokeTimeout,
		reg:           reg,
	}
}

func (d *Dispatcher) getOrCreate(name string) *FunctionExecutor {
	d.mu.RLock()
	ex := d.executors[name]
	d.mu.RUnlock()
	if ex != nil {
		return ex
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if ex = d.executors[name]; ex != nil {
		return ex
	}
	ex = NewFunctionExecutor(name, d.reg, d.queueSize, d.workers, d.globalLimiter)
	d.executors[name] = ex
	return ex
}

func (d *Dispatcher) Invoke(name string, input map[string]any) (map[string]any, error) {
	ex := d.getOrCreate(name)
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	return ex.Enqueue(ctx, input)
}
