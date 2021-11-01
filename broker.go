package asynctask

import (
	"time"

	"github.com/go-redis/redis/v8"
)

// See rdb.RDB as a reference implementation.
type Broker interface {
	Ping() error
	Enqueue(msg *TaskMessage) error
	EnqueueUnique(msg *TaskMessage, ttl time.Duration) error
	Dequeue(qnames ...string) (*TaskMessage, time.Time, error)
	Done(msg *TaskMessage) error
	Requeue(msg *TaskMessage) error
	Schedule(msg *TaskMessage, processAt time.Time) error
	ScheduleUnique(msg *TaskMessage, processAt time.Time, ttl time.Duration) error
	Retry(msg *TaskMessage, processAt time.Time, errMsg string, isFailure bool) error
	Archive(msg *TaskMessage, errMsg string) error
	ForwardIfReady(qnames ...string) error
	ListDeadlineExceeded(deadline time.Time, qnames ...string) ([]*TaskMessage, error)
	WriteServerState(info *ServerInfo, workers []*WorkerInfo, ttl time.Duration) error
	ClearServerState(host string, pid int, serverID string) error
	CancelationPubSub() (*redis.PubSub, error) // TODO: Need to decouple from redis to support other brokers
	PublishCancelation(id string) error
	Close() error
}
