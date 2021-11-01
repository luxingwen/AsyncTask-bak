package asynctask

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
)

type Error struct {
	Code Code
	Op   Op
	Err  error
}

func (e *Error) DebugString() string {
	var b strings.Builder
	if e.Op != "" {
		b.WriteString(string(e.Op))
	}
	if e.Code != Unspecified {
		if b.Len() > 0 {
			b.WriteString(": ")
		}
		b.WriteString(e.Code.String())
	}
	if e.Err != nil {
		if b.Len() > 0 {
			b.WriteString(": ")
		}
		b.WriteString(e.Err.Error())
	}
	return b.String()
}

func (e *Error) Error() string {
	var b strings.Builder
	if e.Code != Unspecified {
		b.WriteString(e.Code.String())
	}
	if e.Err != nil {
		if b.Len() > 0 {
			b.WriteString(": ")
		}
		b.WriteString(e.Err.Error())
	}
	return b.String()
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Code defines the canonical error code.
type Code uint8

// List of canonical error codes.
const (
	Unspecified Code = iota
	NotFound
	FailedPrecondition
	Internal
	AlreadyExists
	Unknown
	// Note: If you add a new value here, make sure to update String method.
)

func (c Code) String() string {
	switch c {
	case Unspecified:
		return "ERROR_CODE_UNSPECIFIED"
	case NotFound:
		return "NOT_FOUND"
	case FailedPrecondition:
		return "FAILED_PRECONDITION"
	case Internal:
		return "INTERNAL_ERROR"
	case AlreadyExists:
		return "ALREADY_EXISTS"
	case Unknown:
		return "UNKNOWN"
	}
	panic(fmt.Sprintf("unknown error code %d", c))
}

var (
	// ErrNoProcessableTask indicates that there are no tasks ready to be processed.
	ErrNoProcessableTask = errors.New("no tasks are ready for processing")

	// ErrDuplicateTask indicates that another task with the same unique key holds the uniqueness lock.
	ErrDuplicateTask = errors.New("task already exists")
)

// Op describes an operation, usually as the package and method,
// such as "rdb.Enqueue".
type Op string

func Err(args ...interface{}) error {
	if len(args) == 0 {
		panic("call to errors.E with no arguments")
	}
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			e.Op = arg
		case Code:
			e.Code = arg
		case error:
			e.Err = arg
		case string:
			e.Err = errors.New(arg)
		default:
			_, file, line, _ := runtime.Caller(1)
			log.Printf("errors.E: bad call from %s:%d: %v", file, line, args)
			return fmt.Errorf("unknown type %T, value %v in error call", arg, arg)
		}
	}
	return e
}
