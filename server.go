package asynctask

import (
	"context"
	"time"
)

type ServerInfo struct {
	Host              string
	PID               int
	ServerID          string
	Concurrency       int
	Queues            map[string]int
	StrictPriority    bool
	Status            string
	Started           time.Time
	ActiveWorkerCount int
}

type Handler interface {
	ProcessTask(context.Context, *Task) error
}

type HandlerFunc func(context.Context, *Task) error

func (fn HandlerFunc) ProcessTask(ctx context.Context, task *Task) error {
	return fn(ctx, task)
}
