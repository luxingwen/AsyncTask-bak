package asynctask

import (
	"sync"
	"time"
)

// 任务转发

type forwarder struct {
	broker Broker

	// channel to communicate back to the long running "forwarder" goroutine.
	done chan struct{}

	// list of queue names to check and enqueue.
	queues []string

	// poll interval on average
	avgInterval time.Duration
}

type forwarderParams struct {
	broker   Broker
	queues   []string
	interval time.Duration
}

func newForwarder(params forwarderParams) *forwarder {
	return &forwarder{
		broker:      params.broker,
		done:        make(chan struct{}),
		queues:      params.queues,
		avgInterval: params.interval,
	}
}

func (f *forwarder) shutdown() {
	f.done <- struct{}{}
}

// start starts the "forwarder" goroutine.
func (f *forwarder) start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-f.done:
				return
			case <-time.After(f.avgInterval):
				f.exec()
			}
		}
	}()
}

func (f *forwarder) exec() {
	if err := f.broker.ForwardIfReady(f.queues...); err != nil {
		//f.logger.Errorf("Could not enqueue scheduled tasks: %v", err)
	}
}
