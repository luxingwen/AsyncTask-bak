package asynctask

import "time"

type Task struct {
	typename string

	payload []byte

	doneCall []Task
}

func (t *Task) Type() string    { return t.typename }
func (t *Task) Payload() []byte { return t.payload }

func (t *Task) DoneCall() []Task {
	return t.doneCall
}

// NewTask returns a new Task given a type name and payload data.
func NewTask(typename string, payload []byte) *Task {
	return &Task{
		typename: typename,
		payload:  payload,
	}
}

type TaskInfo struct {
	ID string

	Name string

	State TaskState

	Payload []byte

	Deadline time.Time

	DoneCall []*TaskMessage
}

type TaskState int

const (
	TaskStateActive TaskState = iota + 1

	TaskStateDoing

	TaskStateDone

	TaskStateScheduled

	TaskStateArchived
)

func (s TaskState) String() string {
	switch s {
	case TaskStateActive:
		return "active"
	case TaskStateDoing:
		return "doing"
	case TaskStateDone:
		return "done"
	case TaskStateScheduled:
		return "scheduled"
	case TaskStateArchived:
		return "archived"
	}
	panic("async task: unknown task state")
}

type TaskMessage struct {
	Typename string // 任务类型

	ID string

	Name string

	State TaskState

	Payload []byte

	Deadline int64
}

type WorkerInfo struct {
	Host     string
	PID      int
	ServerID string
	ID       string
	Type     string
	Payload  []byte
	Queue    string
	Started  time.Time
	Deadline time.Time
}
