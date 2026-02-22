// Package undo provides a lightweight command-based undo stack for the TUI.
package undo

import "sync"

// Operation holds a pair of functions to apply and reverse a mutation.
type Operation struct {
	// Description is a human-readable label shown in the status bar.
	Description string
	// Undo reverses the operation (e.g., restore previous status).
	Undo func() error
	// Redo re-applies the operation after an undo.
	Redo func() error
}

// Stack is a thread-safe undo history.
type Stack struct {
	mu  sync.Mutex
	ops []Operation
	pos int // points to the next "undo" slot (len-1 is top)
	max int
}

// New creates a stack with the given max depth.
func New(max int) *Stack {
	if max <= 0 {
		max = 50
	}
	return &Stack{max: max}
}

// Push records a new operation. It clears any redo history beyond the
// current position and enforces the max depth.
func (s *Stack) Push(op Operation) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Truncate redo branch
	s.ops = s.ops[:s.pos]
	s.ops = append(s.ops, op)

	// Enforce max
	if len(s.ops) > s.max {
		s.ops = s.ops[len(s.ops)-s.max:]
	}
	s.pos = len(s.ops)
}

// Undo pops and reverses the most recent operation.
// Returns ("", nil) if the stack is empty.
func (s *Stack) Undo() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pos == 0 {
		return "", nil
	}
	s.pos--
	op := s.ops[s.pos]
	err := op.Undo()
	return op.Description, err
}

// Redo re-applies the most recently undone operation.
func (s *Stack) Redo() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pos >= len(s.ops) {
		return "", nil
	}
	op := s.ops[s.pos]
	s.pos++
	err := op.Redo()
	return op.Description, err
}

// CanUndo reports whether there is an operation to undo.
func (s *Stack) CanUndo() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pos > 0
}

// CanRedo reports whether there is an operation to redo.
func (s *Stack) CanRedo() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pos < len(s.ops)
}

// Len returns the total number of operations stored.
func (s *Stack) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.ops)
}
