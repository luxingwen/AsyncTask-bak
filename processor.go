package asynctask

import (
	"errors"
	"sync"
	"time"
)

type processor struct {
	broker Broker

	handler Handler

	doneHandler Handler

	queues []string

	queueConfig map[string]int

	isFailureFunc func(error) bool

	// channel via which to send sync requests to syncer.
	syncRequestCh chan<- *syncRequest

	// sema is a counting semaphore to ensure the number of active workers
	// does not exceed the limit.
	sema chan struct{}

	// channel to communicate back to the long running "processor" goroutine.
	// once is used to send value to the channel only once.
	done chan struct{}
	once sync.Once

	// quit channel is closed when the shutdown of the "processor" goroutine starts.
	quit chan struct{}

	// abort channel communicates to the in-flight worker goroutines to stop.
	abort chan struct{}

	finished chan<- *TaskMessage

	shutdownTimeout time.Duration
}

// Note: stops only the "processor" goroutine, does not stop workers.
// It's safe to call this method multiple times.
func (p *processor) stop() {
	p.once.Do(func() {
		// Unblock if processor is waiting for sema token.
		close(p.quit)
		// Signal the processor goroutine to stop processing tasks
		// from the queue.
		p.done <- struct{}{}
	})
}

// NOTE: once shutdown, processor cannot be re-started.
func (p *processor) shutdown() {
	p.stop()

	time.AfterFunc(p.shutdownTimeout, func() { close(p.abort) })

	// block until all workers have released the token
	for i := 0; i < cap(p.sema); i++ {
		p.sema <- struct{}{}
	}

}

func (p *processor) start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-p.done:
				return
			default:
				p.exec()
			}
		}
	}()
}

// exec pulls a task out of the queue and starts a worker goroutine to
// process the task.
func (p *processor) exec() {
	select {
	case <-p.quit:
		return
	case p.sema <- struct{}{}: // acquire token
		qnames := p.queues
		msg, deadline, err := p.broker.Dequeue(qnames...)
		switch {
		case errors.Is(err, ErrNoProcessableTask):
			time.Sleep(time.Second)
			<-p.sema // release token
			return
		case err != nil:
			<-p.sema // release token
			return
		}

		p.starting <- &workerInfo{msg, time.Now(), deadline}
		go func() {
			defer func() {
				p.finished <- msg
				<-p.sema // release token
			}()

			ctx, cancel := createContext(msg, deadline)
			p.cancelations.Add(msg.ID.String(), cancel)
			defer func() {
				cancel()
				p.cancelations.Delete(msg.ID.String())
			}()

			// check context before starting a worker goroutine.
			select {
			case <-ctx.Done():
				// already canceled (e.g. deadline exceeded).
				p.retryOrArchive(ctx, msg, ctx.Err())
				return
			default:
			}

			resCh := make(chan error, 1)
			go func() {
				resCh <- p.perform(ctx, NewTask(msg.Type, msg.Payload))
			}()

			select {
			case <-p.abort:
				// time is up, push the message back to queue and quit this worker goroutine.
				//p.logger.Warnf("Quitting worker. task id=%s", msg.ID)
				p.requeue(msg)
				return
			case <-ctx.Done():
				return
			case resErr := <-resCh:
				// Note: One of three things should happen.
				// 1) Done     -> Removes the message from Active
				// 2) Retry    -> Removes the message from Active & Adds the message to Retry
				// 3) Archive  -> Removes the message from Active & Adds the message to archive
				if resErr != nil {
					return
				}
				p.markAsDone(ctx, msg)
			}
		}()
	}
}

func (p *processor) requeue(msg *TaskMessage) {
	err := p.broker.Requeue(msg)
	if err != nil {
		//p.logger.Errorf("Could not push task id=%s back to queue: %v", msg.ID, err)
	} else {
		//p.logger.Infof("Pushed task id=%s back to queue", msg.ID)
	}
}
