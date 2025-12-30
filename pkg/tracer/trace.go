// Package tracer provides execution tracing capabilities.
package tracer

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Trace represents a test execution trace.
type Trace struct {
	TestName  string    `json:"test_name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Calls     []Call    `json:"calls"`
	mu        sync.Mutex
}

// Call represents a function call event.
type Call struct {
	Event     string    `json:"event"`
	Function  string    `json:"function"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
	Depth     int       `json:"depth"`
}

var (
	currentTrace *Trace
	traceMu      sync.Mutex
	callDepth    int
)

// StartTrace begins a new trace for a test.
func StartTrace(testName string) *Trace {
	traceMu.Lock()
	defer traceMu.Unlock()

	currentTrace = &Trace{
		TestName:  testName,
		StartTime: time.Now(),
		Calls:     []Call{},
	}
	callDepth = 0

	return currentTrace
}

// StopTrace stops the current trace and returns it.
func StopTrace() *Trace {
	traceMu.Lock()
	defer traceMu.Unlock()

	trace := currentTrace
	currentTrace = nil
	callDepth = 0

	return trace
}

// Enter records entry into a function.
func Enter(fn string) {
	traceMu.Lock()
	defer traceMu.Unlock()

	if currentTrace == nil {
		return
	}

	currentTrace.mu.Lock()
	defer currentTrace.mu.Unlock()

	currentTrace.Calls = append(currentTrace.Calls, Call{
		Event:     "enter",
		Function:  fn,
		Timestamp: time.Now(),
		Depth:     callDepth,
	})

	callDepth++
}

// ExitSuccess records successful exit from a function.
func ExitSuccess(fn string) {
	traceMu.Lock()
	defer traceMu.Unlock()

	if currentTrace == nil {
		return
	}

	callDepth--
	if callDepth < 0 {
		callDepth = 0
	}

	currentTrace.mu.Lock()
	defer currentTrace.mu.Unlock()

	currentTrace.Calls = append(currentTrace.Calls, Call{
		Event:     "exit_success",
		Function:  fn,
		Timestamp: time.Now(),
		Depth:     callDepth,
	})
}

// ExitError records exit from a function with an error.
func ExitError(fn string, err error) {
	traceMu.Lock()
	defer traceMu.Unlock()

	if currentTrace == nil {
		return
	}

	callDepth--
	if callDepth < 0 {
		callDepth = 0
	}

	currentTrace.mu.Lock()
	defer currentTrace.mu.Unlock()

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	currentTrace.Calls = append(currentTrace.Calls, Call{
		Event:     "exit_error",
		Function:  fn,
		Timestamp: time.Now(),
		Error:     errMsg,
		Depth:     callDepth,
	})
}

// Exit records exit from a function, routing to ExitSuccess or ExitError.
func Exit(fn string, err error) {
	if err != nil {
		ExitError(fn, err)
	} else {
		ExitSuccess(fn)
	}
}

// Save saves the trace to a JSON file.
func (t *Trace) Save(filename string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.EndTime = time.Now()

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize trace: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return fmt.Errorf("failed to write trace file: %w", err)
	}

	return nil
}

// Save saves the current trace to a JSON file.
func Save(filename string) error {
	traceMu.Lock()
	trace := currentTrace
	traceMu.Unlock()

	if trace == nil {
		return fmt.Errorf("no active trace")
	}

	return trace.Save(filename)
}
