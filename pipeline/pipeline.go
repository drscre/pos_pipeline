package pipeline

import (
	"context"
	"database/sql"
	"time"
)

// Declaration of pipeline (pipeline type)
type Pipeline struct {
	steps       []IStep
	retryPolicy interface{} // default retry policy for steps

	storage stateStorage
}

// State of the concrete pipeline instance
// @TODO: execution context: "registers" (data) & instruction pointer
type State struct {
	ID                string      // used to select, lock and update data
	Name              string      // pipeline name
	Data              interface{} // data context
	LastCompletedStep string
	Error             string // Non-empty if pipeline was terminated with non-recoverable error
}

type IStep interface {
	name() string
	isStep()
}

type SimpleStep struct {
	n string
	h     func(data interface{}) (modifiedData interface{}, err error)
	delay time.Duration
}

func (s *SimpleStep) Name(name string) *SimpleStep {
	s.n = name
	return s
}
func (s SimpleStep) name() string { return s.n }
func (SimpleStep) isStep()      {}

func (s *SimpleStep) Delayed(delay time.Duration) *SimpleStep {
	s.delay = delay
	return s
}

func (s *SimpleStep) RetryPolicy() *SimpleStep {
	return s
}

func Step(h func(data interface{}) (modifiedData interface{}, err error)) *SimpleStep {
	return &SimpleStep{
		h: h,
	}
}

// cond is not supposed to have side-effects (while technically it can)
// @TODO: should work with nested ifs
func If(cond func(data interface{}) bool, step IStep, more ...IStep) *IfStep {
	return &IfStep{}
}

type IfStep struct {
}

func (IfStep) name() string { return "" }
func (IfStep) isStep()      {}

func While(cond func(data interface{}) bool, step IStep, more ...IStep) *WhileStep {
	return &WhileStep{}
}

type WhileStep struct {
	delay time.Duration
}

func (WhileStep) name() string { return "" }
func (WhileStep) isStep()      {}

func (s *WhileStep) Delayed(delay time.Duration) *WhileStep {
	s.delay = delay
	return s
}

func Sleep(duration time.Duration) *SleepStep {
	return &SleepStep{duration: duration}
}

type SleepStep struct {
	duration time.Duration
}

func (SleepStep) name() string { return "" }
func (SleepStep) isStep()      {}

func Retry(err error) error {
	return err
}

type EventLoopStep struct {
}

func (EventLoopStep) name() string { return "" }
func (EventLoopStep) isStep()      {}

type EventLoopListener struct{}

func EventLoop(cond func(data interface{}) bool, eventListeners ...*EventLoopListener) *EventLoopStep {
	return &EventLoopStep{}
}

func OnEvent(func(data interface{}, event interface{}) (modifiedData interface{}, err error)) *EventLoopListener {
	return &EventLoopListener{}
}

func OnTimerTick(interval time.Duration, cb func(data interface{}) (modifiedData interface{}, err error)) *EventLoopListener {
	return &EventLoopListener{}
}

func ErrorHandler(h func(data interface{}, pipelineErr error) (modifiedData interface{}, err error)) *ErrorHandlerStep {
	return &ErrorHandlerStep{}
}

type ErrorHandlerStep struct{}

func (ErrorHandlerStep) name() string { return "" }
func (ErrorHandlerStep) isStep()      {}

// Simple step. Proceed to next step if completes without error.
// Retry (with updating data!) on error according to step retry policy.
func (p *Pipeline) Step(name string, cb func(data interface{}) (modifiedData interface{}, err error)) {
	// validate name unique
	p.steps = append(p.steps, simpleStep{
		n:  name,
		cb: cb,
	})
}

// Event loop. Repeatedly accept events until "stop=true" is returned
func (p *Pipeline) Wait(name string, cb func(data interface{}, event interface{}) (modifiedData interface{}, stop bool, err error)) {
	// validate name unique
	p.steps = append(p.steps, eventLoopStep{
		n:  name,
		cb: cb,
	})
}

func Declare(storage stateStorage) Pipeline {
	return Pipeline{
		storage: storage,
	}
}

// Tolerates concurrent pipeline execution - steps might interleave
func (p Pipeline) Run(ctx context.Context, id string) error {
	for {
		tx, _ := p.storage.BeginTx(ctx, nil)
		state, _ := p.storage.SelectForUpdate(tx, id) // also serves as lock to prevent concurrent step execution

		if state.Error != "" { // pipelined already finished with error
			_ = tx.Commit()
			return nil
		}

		step := p.nextStep(state.LastCompletedStep)
		if step == nil {
			_ = tx.Commit()
			return nil // pipeline finished succesfully
		}

		switch step := step.(type) {
		case simpleStep:
			_, err := step.cb(state.Data)
			if err != nil {

				return err // @TODO: go for retry or error on non-recoverable error
			}
		}

		_ = tx.Commit()
	}
}

func (p Pipeline) nextStep(lastCompletedStep string) IStep {
	if lastCompletedStep == "" {
		return p.steps[0]
	}
	for i, step := range p.steps {
		if step.name() == lastCompletedStep {
			if i+1 < len(p.steps) {
				return p.steps[i+1]
			} else {
				return nil // indicates last step
			}
		}
	}
	panic("broken pipeline")
}

type simpleStep struct {
	n  string
	cb func(data interface{}) (modifiedData interface{}, err error)
}

func (s simpleStep) name() string {
	return s.n
}
func (simpleStep) isStep() {}

type eventLoopStep struct {
	n  string
	cb func(data interface{}, event interface{}) (modifiedData interface{}, stop bool, err error)
}

func (s eventLoopStep) name() string { return s.n }
func (eventLoopStep) isStep()        {}

// Storage for concrete pipeline execution state
type stateStorage interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	SelectForUpdate(conn DBConn, id string) (State, error)
	Update(conn DBConn, id string, newData interface{}, completedStep string) error
}

type DBConn interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
