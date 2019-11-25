// Copyright (c) 2013 - Max Persson <max@looplab.se>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package fsm implements a finite state machine.
//
// It is heavily based on two FSM implementations:
//
// Javascript Finite State Machine
// https://github.com/jakesgordon/javascript-state-machine
//
// Fysom for Python
// https://github.com/oxplot/fysom (forked at https://github.com/mriehl/fysom)
//

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "EventType=string StateType=string"

package fsm

import (
	"github.com/cheekybits/genny/generic"
	"sync"
)

// EventTypeEventStateTypeStateTransitioner is an interface for the Finite State Machine's transition function.
type EventTypeEventStateTypeStateTransitioner interface {
	Transition(*EventTypeStateTypeFiniteStateMachine) error
}

// EventTypeStateTypeFiniteStateMachine is the state machine that holds the current state.
//
// It has to be created with NewFSM to function properly.
type EventTypeStateTypeFiniteStateMachine struct {

	//
	// BeforeEvent called before all events
	BeforeEvent TransitionFunc

	//
	// LeaveState called before leaving all states
	LeaveState TransitionFunc

	//
	// EnterState called after entering all states
	EnterState TransitionFunc

	//
	// AfterEvent called after all events
	AfterEvent TransitionFunc

	// current is the state that the EventTypeStateTypeFiniteStateMachine is currently in.
	current StateType

	// transitions maps events and source states to destination states.
	transitions map[eKey]StateType

	// callbacks maps events and tragers to callback functions.
	callbacks map[cKey]TransitionFunc

	// transition is the internal transition functions used either directly
	// or when Transition is called in an asynchronous state transition.
	transition func() error
	// transitionerObj calls the FSM's transition() function.
	transitionerObj EventTypeEventStateTypeStateTransitioner

	// stateMu guards access to the current state.
	stateMu sync.RWMutex
	// eventMu guards access to Event() and Event().
	eventMu sync.Mutex
}

type StateType generic.Type

type StateTypeStates []StateType

type EventType generic.Type

// Event represents an event when initializing the EventTypeStateTypeFiniteStateMachine.
//
// The event can have one or more source states that is valid for performing
// the transition. If the FSM is in one of the source states it will end up in
// the specified destination state, calling all defined callbacks as it goes.
type EventTypeEvent struct {
	// Name is the event name used when calling for a transition.
	Label EventType

	// Src is the source states that the EventTypeStateTypeFiniteStateMachine must be in to perform a
	// state transition.
	Src StateType

	// Dst is the destination state that the EventTypeStateTypeFiniteStateMachine will be in to perform the transition.
	Dst StateType

	// 1. before_<EVENT> - called before event named <EVENT>
	//
	BeforeEvent TransitionFunc

	// 3. leave_<OLD_STATE> - called before leaving <OLD_STATE>
	//
	LeaveState TransitionFunc

	// 5. enter_<NEW_STATE> - called after entering <NEW_STATE>
	//
	// 1. <NEW_STATE> - called after entering <NEW_STATE>
	//
	EnterState TransitionFunc

	// 7. after_<EVENT> - called after event named <EVENT>
	//
	// 2. <EVENT> - called after event named <EVENT>
	//
	AfterEvent TransitionFunc

}

// Events is a shorthand for defining the transition map in NewFSM.
type EventTypeEvents []EventTypeEvent

// NewFSM constructs a EventTypeStateTypeFiniteStateMachine from events and callbacks.
//
// The events and transitions are specified as a slice of Transition struct
// specified as Events. Each Transition is mapped to one or more internal
// transitions from Transition.Src to Transition.Dst.
//
// Callbacks are added as a map specified as Callbacks where the key is parsed
// as the callback event as follows, and called in the same order:
//
// 1. before_<EVENT> - called before event named <EVENT>
//
// 2. before_event - called before all events
//
// 3. leave_<OLD_STATE> - called before leaving <OLD_STATE>
//
// 4. leave_state - called before leaving all states
//
// 5. enter_<NEW_STATE> - called after entering <NEW_STATE>
//
// 6. enter_state - called after entering all states
//
// 7. after_<EVENT> - called after event named <EVENT>
//
// 8. after_event - called after all events
//
// There are also two short form versions for the most commonly used callbacks.
// They are simply the name of the event or state:
//
// 1. <NEW_STATE> - called after entering <NEW_STATE>
//
// 2. <EVENT> - called after event named <EVENT>
//
// If both a shorthand version and a full version is specified it is undefined
// which version of the callback will end up in the internal map. This is due
// to the pseudo random nature of Go maps. No checking for multiple keys is
// currently performed.
func NewEventTypeStateTypeFiniteStateMachine(initial StateType, events EventTypeEvents) *EventTypeStateTypeFiniteStateMachine {
	f := &EventTypeStateTypeFiniteStateMachine{
		transitionerObj: &defaultTransitioner{},
		current:         initial,
		transitions:     make(map[eKey]StateType),
		callbacks:       make(map[cKey]TransitionFunc),
	}

	// Build transition map and store sets of all events and states.
	for _, e := range events {
		src := e.Src
		f.transitions[eKey{e.Label, src}] = e.Dst
		if e.BeforeEvent != nil {
			f.callbacks[cKey{e.Label, callbackBeforeEvent}] = e.BeforeEvent
		}
		if e.LeaveState != nil {
			f.callbacks[cKey{ e.Src, callbackLeaveState}] = e.LeaveState
		}
		if e.EnterState != nil {
			f.callbacks[cKey{ e.Dst, callbackEnterState}] = e.EnterState
		}
		if e.AfterEvent != nil {
			f.callbacks[cKey{e.Label, callbackAfterEvent}] = e.AfterEvent
		}

	}

	return f
}

// Current returns the current state of the EventTypeStateTypeFiniteStateMachine.
func (f *EventTypeStateTypeFiniteStateMachine) Current() StateType {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	return f.current
}

// Is returns true if state is the current state.
func (f *EventTypeStateTypeFiniteStateMachine) Is(state StateType) bool {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	return state == f.current
}

// State allows the user to move to the given state from current state.
// The call does not trigger any callbacks, if defined.
func (f *EventTypeStateTypeFiniteStateMachine) State(state StateType) {
	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	f.current = state
	return
}

// Can returns true if event can occur in the current state.
func (f *EventTypeStateTypeFiniteStateMachine) Can(event EventType) bool {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	_, ok := f.transitions[eKey{event, f.current}]
	return ok && (f.transition == nil)
}

// AvailableTransitions returns a list of transitions available in the
// current state.
func (f *EventTypeStateTypeFiniteStateMachine) AvailableTransitions() StateTypeStates {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	var transitions StateTypeStates
	for key := range f.transitions {
		if key.src == f.current {
			transitions = append(transitions, key.event)
		}
	}
	return transitions
}

// Cannot returns true if event can not occur in the current state.
// It is a convenience method to help code read nicely.
func (f *EventTypeStateTypeFiniteStateMachine) Cannot(event EventType) bool {
	return !f.Can(event)
}

// Event initiates a state transition with the named event.
//
// The call takes a variable number of arguments that will be passed to the
// callback, if defined.
//
// It will return nil if the state change is ok or one of these errors:
//
// - event X inappropriate because previous transition did not complete
//
// - event X inappropriate in current state Y
//
// - event X does not exist
//
// - internal error on state transition
//
// The last error should never occur in this situation and is a sign of an
// internal bug.
func (f *EventTypeStateTypeFiniteStateMachine) Event(event EventType, args ...interface{}) error {
	f.eventMu.Lock()
	defer f.eventMu.Unlock()

	f.stateMu.RLock()
	defer f.stateMu.RUnlock()

	if f.transition != nil {
		return InTransitionError{event}
	}

	dst, ok := f.transitions[eKey{event, f.current}]
	if !ok {
		for ekey := range f.transitions {
			if ekey.event == event {
				return InvalidEventError{event, f.current}
			}
		}
		return UnknownEventError{event}
	}

	t := new(cancelTransition)
	t.event = event
	t.src = f.current
	t.dst = dst
	t.args = args

	err := f.beforeEventCallbacks(t)
	if err != nil {
		return err
	}

	if f.current == dst {
		err := f.afterEventCallbacks(t)
		if err != nil {
			return NoTransitionError{err}
		}
		return NoTransitionError{t.Err()}
	}

	// Setup the transition, call it later.
	f.transition = func() error {
		f.stateMu.Lock()
		f.current = dst
		f.stateMu.Unlock()

		if err := f.enterStateCallbacks(t); err != nil {
			return err
		}
		if err := f.afterEventCallbacks(t); err != nil {
			return err
		}
		return nil
	}

	if err = f.leaveStateCallbacks(t); err != nil {
		if err == Canceled {
			f.transition = nil
		}
		return err
	}

	// Perform the rest of the transition, if not asynchronous.
	f.stateMu.RUnlock()
	err = f.transitionerObj.Transition(f)
	f.stateMu.RLock()
	if err != nil {
		return err
	}

	return t.Err()
}

// Transition wraps transitioner.transition.
func (f *EventTypeStateTypeFiniteStateMachine) Transition() error {
	f.eventMu.Lock()
	defer f.eventMu.Unlock()
	return f.transitionerObj.Transition(f)
}

// beforeEventCallbacks calls the before_ callbacks, first the named then the
// general version.
func (f *EventTypeStateTypeFiniteStateMachine) beforeEventCallbacks(t Transition) error {
	event := t.Event()
	if fn, ok := f.callbacks[cKey{event, callbackBeforeEvent}]; ok {
		err := fn(t)
		if err != nil {
			return err
		}
		if t.Err() != nil {
			return t.Err()
		}
	}
	if f.BeforeEvent != nil {
		err := f.BeforeEvent(t)
		if err != nil {
			return err
		}
		if t.Err() != nil {
			return t.Err()
		}
	}
	return nil
}

// leaveStateCallbacks calls the leave_ callbacks, first the named then the
// general version.
func (f *EventTypeStateTypeFiniteStateMachine) leaveStateCallbacks(t Transition) error {
	if fn, ok := f.callbacks[cKey{f.current, callbackLeaveState}]; ok {
		if err := fn(t); err != nil {
			return err
		}
		if t.Err() != nil {
			return t.Err()
		} else if t.Async() {
			return AsyncError{t.Err()}
		}
	}
	if f.LeaveState != nil {
		if err := f.LeaveState(t); err != nil {
			return err
		}
		if t.Err() != nil {
			return t.Err()
		} else if t.Async() {
			return AsyncError{t.Err()}
		}
	}
	return nil
}

// enterStateCallbacks calls the enter_ callbacks, first the named then the
// general version.
func (f *EventTypeStateTypeFiniteStateMachine) enterStateCallbacks(t Transition) error {
	if fn, ok := f.callbacks[cKey{f.current, callbackEnterState}]; ok {
		if err := fn(t); err != nil {
			return err
		}
	}
	if f.EnterState != nil {
		if err := f.EnterState(t); err != nil {
			return err
		}
	}
	return nil
}

// afterEventCallbacks calls the after_ callbacks, first the named then the
// general version.
func (f *EventTypeStateTypeFiniteStateMachine) afterEventCallbacks(t Transition) error {
	if fn, ok := f.callbacks[cKey{t.Event(), callbackAfterEvent}]; ok {
		if err := fn(t); err != nil {
			return err
		}
	}
	if f.AfterEvent != nil {
		if err := f.AfterEvent(t); err != nil {
			return err
		}
	}
	return nil
}

