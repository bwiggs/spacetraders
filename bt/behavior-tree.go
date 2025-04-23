package bt

import (
	"fmt"
	"log/slog"
	"reflect"
)

type Blackboard = any

// BehaviorStatus represents the status of a behavior node.
type BehaviorStatus int

const (
	// Success indicates that the behavior node has succeeded.
	Success BehaviorStatus = iota
	// Failure indicates that the behavior node has failed.
	Failure
	// Running indicates that the behavior node is still running.
	Running
)

// BehaviorNode defines the interface for behavior tree nodes.
type BehaviorNode interface {
	Tick(Blackboard) BehaviorStatus
}

func printResult(node any, status BehaviorStatus) {
	return
	st := "Success"
	switch status {
	case 1:
		st = "failure"
	case 2:
		st = "running"
	}
	slog.Debug(fmt.Sprintf("%s: %s", reflect.TypeOf(node).String(), st))
}

// Sequence is a behavior node that executes its children in sequence.
type Sequence struct {
	children []BehaviorNode
}

// NewSequence creates a new Sequence node with the given children.
// A sequence will fail if any child fails.
func NewSequence(children ...BehaviorNode) *Sequence {
	return &Sequence{children: children}
}

// Tick executes each child node in sequence until one fails.
func (s *Sequence) Tick(bb any) BehaviorStatus {
	for _, child := range s.children {
		status := child.Tick(bb)
		printResult(child, status)

		if status != Success {
			return status
		}
	}
	return Success
}

// Selector is a behavior node that executes its children until one succeeds.
type Selector struct {
	children []BehaviorNode
}

// NewSelector creates a new Selector node with the given children.
func NewSelector(children ...BehaviorNode) *Selector {
	return &Selector{children: children}
}

// NewOr creates a new Selector node with the given children.
func NewOr(children ...BehaviorNode) *Selector {
	return &Selector{children: children}
}

// Tick executes each child node until one succeeds.
func (s *Selector) Tick(bb any) BehaviorStatus {
	for _, child := range s.children {
		status := child.Tick(bb)
		// printResult(child, status)
		if status != Failure {
			return status
		}
	}
	return Failure
}

type Inversion struct {
	child BehaviorNode
}

func Not(child BehaviorNode) *Inversion {
	return &Inversion{child}
}

// Tick executes each child node until one succeeds.
func (i *Inversion) Tick(bb any) BehaviorStatus {
	r := i.child.Tick(bb)
	if r == Success {
		return Failure
	} else if r == Failure {
		return Success
	}
	return Running
}
