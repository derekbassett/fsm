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

package fsm

import (
	"errors"
	"sync"
)

// Transition carries a signal for cancelling.
type Transition interface {
	// Async is Set if the transition should be asynchronous
	Async() bool

	// SetAsync can be called in leave_<STATE> to do an asynchronous state transition.
	//
	// The current state transition will be on hold in the old state until a final
	// call to Transition is made. This will complete the transition and possibly
	// call the other callbacks.
	SetAsync()

	// Cancel can be called in before_<EVENT> or leave_<STATE> to cancel the
	// current transition before it happens. It takes an optional error, which will
	// overwrite Err if set before.
	Cancel()

	// Event is the event that generated this transition.
	Event() EventType

	// Src is the state before the transition.
	Src() StateType

	// Dst is the state after the transition.
	Dst() StateType

	// Err is an optional error that can be returned from a callback.
	Err() error

	// Args is a list of arguments
	Args() []interface{}
}

// Callback is a function type that callbacks should use. Transition is the current
// event info as the callback happens.
type TransitionFunc func(Transition) error

// Callbacks is a shorthand for defining the callbacks in NewFSM.
type Transitions map[string]TransitionFunc

// Canceled is the error returned by Context.Err when the context is canceled.
var Canceled = errors.New("transition canceled")

// closedchan is a reusable closed channel.
var closedchan = make(chan struct{})

func init() {
	close(closedchan)
}

var _ Transition = (*cancelTransition)(nil)

type cancelTransition struct {
	mu sync.Mutex      // protects following fields
	event EventType
	src StateType
	dst StateType
	done chan struct{} // created lazily, closed by first cancel call
	err error
	// async is an internal flag set if the transition should be asynchronous
	async bool
	args []interface{}
}

func (c *cancelTransition) Event() EventType {
	c.mu.Lock()
	e := c.event
	c.mu.Unlock()
	return e
}

func (c *cancelTransition) Src() StateType {
	return c.src
}

func (c *cancelTransition) Dst() StateType {
	return c.dst
}

// Cancel can be called in before_<EVENT> or leave_<STATE> to cancel the
// current transition before it happens.
func (c *cancelTransition) Cancel() {
	c.cancel(Canceled)
}

func (c *cancelTransition) cancel(err error) {
	if err == nil {
		panic("context: internal error: missing cancel error")
	}
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return
	}
	c.err = err
	if c.done == nil {
		c.done = closedchan
	} else {
		close(c.done)
	}
	c.mu.Unlock()
}

func (c *cancelTransition) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()
	return err
}

func (c *cancelTransition) Async() bool {
	c.mu.Lock()
	async := c.async
	c.mu.Unlock()
	return async
}

// Async can be called in leave_<STATE> to do an asynchronous state transition.
//
// The current state transition will be on hold in the old state until a final
// call to Transition is made. This will complete the transition and possibly
// call the other callbacks.
func (c *cancelTransition) SetAsync() {
	c.mu.Lock()
	c.async = true
	c.mu.Unlock()
}

func (c *cancelTransition) Args() []interface{} {
	c.mu.Lock()
	args := c.args
	c.mu.Unlock()
	return args
}

// Transition is the info that get passed as a reference in the callbacks.
//type Transition struct {
//
//	// Event is the event.
//	Event EventType
//
//	// Src is the state before the transition.
//	Src StateType
//
//	// Dst is the state after the transition.
//	Dst StateType
//
//	// Err is an optional error that can be returned from a callback.
//	Err error
//
//	// Args is a optional list of arguments passed to the callback.
//	Args []interface{}
//
//	// canceled is an internal flag set if the transition is canceled.
//	canceled bool
//
//	// async is an internal flag set if the transition should be asynchronous
//	async bool
//}


