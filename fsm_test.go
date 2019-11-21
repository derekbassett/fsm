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
	"fmt"
	"sync"
	"testing"
	"time"
)

type fakeTransitionerObj struct {
}

func (t fakeTransitionerObj) Transition(f *EventTypeStateTypeFiniteStateMachine) error {
	return &InternalError{}
}

func TestSameState(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "start"},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestState(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"walking",
		EventTypeEvents{
			{Label: "walk", Src: "start", Dst: "walking"},
		},
	)
	fsm.State("start")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'walking'")
	}
	err := fsm.Event("walk")
	if err != nil {
		t.Error("transition is expected no error")
	}
}

func TestBadTransition(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "running"},
		},
	)
	fsm.transitionerObj = new(fakeTransitionerObj)
	err := fsm.Event("run")
	if err == nil {
		t.Error("bad transition should give an error")
	}
}

func TestInappropriateEvent(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"closed",
		EventTypeEvents{
			{Label: "open", Src: "closed", Dst: "open"},
			{Label: "close", Src: "open", Dst: "closed"},
		},
	)
	err := fsm.Event("close")
	if e, ok := err.(InvalidEventError); !ok && e.Event != "close" && e.State != "closed" {
		t.Error("expected 'InvalidEventError' with correct state and event")
	}
}

func TestInvalidEvent(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"closed",
		EventTypeEvents{
			{Label: "open", Src: "closed", Dst: "open"},
			{Label: "close", Src: "open", Dst: "closed"},
		},
	)
	err := fsm.Event("lock")
	if e, ok := err.(UnknownEventError); !ok && e.Event != "close" {
		t.Error("expected 'UnknownEventError' with correct event")
	}
}

func TestMultipleSources(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"one",
		EventTypeEvents{
			{Label: "first", Src: "one", Dst: "two"},
			{Label: "second", Src: "two", Dst: "three"},
			{Label: "reset", Src: "one", Dst: "one"},
			{Label: "reset", Src: "two", Dst: "one"},
			{Label: "reset", Src: "three", Dst: "one"},
		},
	)

	fsm.Event("first")
	if fsm.Current() != "two" {
		t.Error("expected state to be 'two'")
	}
	fsm.Event("reset")
	if fsm.Current() != "one" {
		t.Error("expected state to be 'one'")
	}
	fsm.Event("first")
	fsm.Event("second")
	if fsm.Current() != "three" {
		t.Error("expected state to be 'three'")
	}
	fsm.Event("reset")
	if fsm.Current() != "one" {
		t.Error("expected state to be 'one'")
	}
}

func TestMultipleEvents(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "first", Src: "start", Dst: "one"},
			{Label: "second", Src: "start", Dst: "two"},
			{Label: "reset", Src: "one", Dst: "reset_one"},
			{Label: "reset", Src: "two", Dst: "reset_two"},
			{Label: "reset", Src: "reset_one", Dst: "start"},
			{Label: "reset", Src: "reset_two", Dst: "start"},
		},
	)

	fsm.Event("first")
	fsm.Event("reset")
	if fsm.Current() != "reset_one" {
		t.Error("expected state to be 'reset_one'")
	}
	fsm.Event("reset")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}

	fsm.Event("second")
	fsm.Event("reset")
	if fsm.Current() != "reset_two" {
		t.Error("expected state to be 'reset_two'")
	}
	fsm.Event("reset")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestGenericCallbacks(t *testing.T) {
	beforeEvent := false
	leaveState := false
	enterState := false
	afterEvent := false

	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{
				Label: "run",
				Src:   "start",
				Dst:   "end",
			},
		},
	)
	fsm.BeforeEvent = func(t Transition) error {
		beforeEvent = true
		return nil
	}
	fsm.LeaveState = func(t Transition) error {
		leaveState = true
		return nil
	}
	fsm.EnterState = func(t Transition) error {
		enterState = true
		return nil
	}
	fsm.AfterEvent = func(t Transition) error {
		afterEvent = true
		return nil
	}

	fsm.Event("run")
	if !(beforeEvent && leaveState && enterState && afterEvent) {
		t.Error("expected all callbacks to be called")
	}
}

func TestSpecificCallbacks(t *testing.T) {
	beforeEvent := false
	leaveState := false
	enterState := false
	afterEvent := false

	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{
				Label: "run",
				Src: "start",
				Dst: "end",
				BeforeEvent:func(t Transition) error {
					beforeEvent = true
					return nil
				},
				LeaveState: func(t Transition) error {
					leaveState = true
					return nil
				},
				EnterState: func(t Transition) error {
					enterState = true
					return nil
				},
				AfterEvent:func(t Transition) error {
					afterEvent = true
					return nil
				},
			},
		},
	)

	fsm.Event("run")
	if !(beforeEvent && leaveState && enterState && afterEvent) {
		t.Error("expected all callbacks to be called")
	}
}

//func TestSpecificCallbacksShortform(t *testing.T) {
//	enterState := false
//	afterEvent := false
//
//	fsm := NewEventTypeStateTypeFiniteStateMachine(
//		"start",
//		EventTypeEvents{
//			{Label: "run", Src: StateTypeStates{"start"}, Dst: "end"},
//		},
//		Transitions{
//			"end": func(t Transition) error {
//				enterState = true
//				return nil
//			},
//			"run": func(t Transition) error {
//				afterEvent = true
//				return nil
//			},
//		},
//	)
//
//	fsm.Event("run")
//	if !(enterState && afterEvent) {
//		t.Error("expected all callbacks to be called")
//	}
//}

func TestBeforeEventWithoutTransition(t *testing.T) {
	beforeEvent := true

	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "dontrun", Src: "start", Dst: "start", BeforeEvent: func(t Transition) error {
				beforeEvent = true
				return nil
			}},
		},
	)

	err := fsm.Event("dontrun")
	if e, ok := err.(NoTransitionError); !ok && e.Err != nil {
		t.Error("expected 'NoTransitionError' without custom error")
	}

	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
	if !beforeEvent {
		t.Error("expected callback to be called")
	}
}

func TestCancelBeforeGenericEvent(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end"},
		},
	)
	fsm.BeforeEvent = func(t Transition) error {
		t.Cancel()
		return nil
	}
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestCancelBeforeSpecificEvent(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{ Label: "run", Src: "start", Dst: "end", BeforeEvent: func(t Transition) error {
				t.Cancel()
				return nil
			}},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestCancelLeaveGenericState(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end", LeaveState: func(t Transition) error {
				t.Cancel()
				return nil
			}},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

//func TestCancelLeaveSpecificState(t *testing.T) {
//	fsm := NewEventTypeStateTypeFiniteStateMachine(
//		"start",
//		EventTypeEvents{
//			{Label: "run", Src: StateTypeStates{"start"}, Dst: "end"},
//		},
//		Transitions{
//			"leave_start": func(t Transition) error {
//				t.Cancel()
//				return nil
//			},
//		},
//	)
//	fsm.Event("run")
//	if fsm.Current() != "start" {
//		t.Error("expected state to be 'start'")
//	}
//}

func TestCancelWithError(t *testing.T) {
	expect := fmt.Errorf("error")
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end", BeforeEvent:  func(t Transition) error {
				t.Cancel()
				return expect
			}},
		},
	)
	err := fsm.Event("run")
	if err != expect {
		t.Error("expected custom error")
	}

	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestAsyncTransitionGenericState(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end", },
		},
	)
	fsm.LeaveState = func(t Transition) error {
		t.SetAsync()
		return nil
	}
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
	fsm.Transition()
	if fsm.Current() != "end" {
		t.Error("expected state to be 'end'")
	}
}

func TestAsyncTransitionSpecificState(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end", LeaveState: func(t Transition) error {
				t.SetAsync()
				return nil
			}},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
	fsm.Transition()
	if fsm.Current() != "end" {
		t.Error("expected state to be 'end'")
	}
}

func TestAsyncTransitionInProgress(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end", LeaveState: func(t Transition) error {
				t.SetAsync()
				return nil
			}},
			{Label: "reset", Src: "end", Dst: "start"},
		},
	)
	fsm.Event("run")
	err := fsm.Event("reset")
	if e, ok := err.(InTransitionError); !ok && e.Event != "reset" {
		t.Error("expected 'InTransitionError' with correct state")
	}
	fsm.Transition()
	fsm.Event("reset")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestAsyncTransitionNotInProgress(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end"},
			{Label: "reset", Src: "end", Dst: "start"},
		},
	)
	err := fsm.Transition()
	if _, ok := err.(NotInTransitionError); !ok {
		t.Error("expected 'NotInTransitionError'")
	}
}

func TestCallbackNoError(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end", AfterEvent: func(t Transition) error {
				return nil
			}},
		},
	)
	e := fsm.Event("run")
	if e != nil {
		t.Error("expected no error")
	}
}

func TestCallbackError(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end", AfterEvent: func(t Transition) error {
				return fmt.Errorf("error")
			}},
		},
	)
	e := fsm.Event("run")
	if e == nil || e.Error() != "error" {
		t.Error("expected error to be 'error'")
	}
}

func TestCallbackArgs(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end", AfterEvent: func(tr Transition) error {
				args := tr.Args()
				if len(args) != 1 {
					t.Error("too few arguments")
				}
				arg, ok := args[0].(string)
				if !ok {
					t.Error("not a string argument")
				}
				if arg != "test" {
					t.Error("incorrect argument")
				}
				return nil
			}},
		},
	)
	fsm.Event("run", "test")
}

func TestNoDeadLock(t *testing.T) {
	var fsm *EventTypeStateTypeFiniteStateMachine
	fsm = NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end", AfterEvent: func(t Transition) error {
				fsm.Current() // Should not result in a panic / deadlock
				return nil
			}},
		},
	)
	fsm.Event("run")
}

func TestThreadSafetyRaceCondition(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end", AfterEvent: func(t Transition) error {
				return nil
			}},
		},
	)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = fsm.Current()
	}()
	fsm.Event("run")
	wg.Wait()
}

func TestDoubleTransition(t *testing.T) {
	var fsm *EventTypeStateTypeFiniteStateMachine
	var wg sync.WaitGroup
	wg.Add(2)
	fsm = NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "end", BeforeEvent: func(tr Transition) error {
				wg.Done()
				// Imagine a concurrent event coming in of the same type while
				// the data access mutex is unlocked because the current transition
				// is running its event callbacks, getting around the "active"
				// transition checks
				if len(tr.Args()) == 0 {
					// Must be concurrent so the test may pass when we add a mutex that synchronizes
					// calls to Transition(...). It will then fail as an inappropriate transition as we
					// have changed state.
					go func() {
						if err := fsm.Event("run", "second run"); err != nil {
							fmt.Println(err)
							wg.Done() // It should fail, and then we unfreeze the test.
						}
					}()
					time.Sleep(20 * time.Millisecond)
				} else {
					panic("Was able to reissue an event mid-transition")
				}
				return nil
			}},
		},
	)
	if err := fsm.Event("run"); err != nil {
		fmt.Println(err)
	}
	wg.Wait()
}

func TestNoTransition(t *testing.T) {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"start",
		EventTypeEvents{
			{Label: "run", Src: "start", Dst: "start"},
		},
	)
	err := fsm.Event("run")
	if _, ok := err.(NoTransitionError); !ok {
		t.Error("expected 'NoTransitionError'")
	}
}

func ExampleNewFSM() {
	beforeEvent := func(t Transition) error {
		fmt.Println("before_event")
		return nil
	}
	beforeWarn := func(t Transition) error {
		fmt.Println("before_warn")
		return nil
	}
	leaveGreen := func(t Transition) error {
		fmt.Println("leave_green")
		return nil
	}
	leaveState := func(t Transition) error {
		fmt.Println("leave_state")
		return nil
	}
	enterYellow := func(t Transition) error {
		fmt.Println("enter_yellow")
		return nil
	}
	enterState := func(t Transition) error {
		fmt.Println("enter_state")
		return nil
	}
	afterWarn := func(t Transition) error {
		fmt.Println("after_warn")
		return nil
	}
	afterEvent := func(t Transition) error {
		fmt.Println("after_event")
		return nil
	}

	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"green",
		EventTypeEvents{
			{Label: "warn", Src: "green", Dst: "yellow", BeforeEvent: beforeWarn, LeaveState: leaveGreen, AfterEvent: afterWarn, EnterState: enterYellow},
			{Label: "panic", Src: "yellow", Dst: "red"},
			{Label: "panic", Src: "green", Dst: "red", LeaveState: leaveGreen},
			{Label: "calm", Src: "red", Dst: "yellow", EnterState: enterYellow},
			{Label: "clear", Src: "yellow", Dst: "green", EnterState: enterYellow },
		},
	)

	fsm.BeforeEvent = beforeEvent
	fsm.LeaveState = leaveState
	fsm.EnterState = enterState
	fsm.AfterEvent = afterEvent

	fmt.Println(fsm.Current())
	err := fsm.Event("warn")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	// Output:
	// green
	// before_warn
	// before_event
	// leave_green
	// leave_state
	// enter_yellow
	// enter_state
	// after_warn
	// after_event
	// yellow
}

func ExampleFSM_Current() {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"closed",
		EventTypeEvents{
			{Label: "open", Src: "closed", Dst: "open"},
			{Label: "close", Src: "open", Dst: "closed"},
		},
	)
	fmt.Println(fsm.Current())
	// Output: closed
}

func ExampleFSM_Is() {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"closed",
		EventTypeEvents{
			{Label: "open", Src: "closed", Dst: "open"},
			{Label: "close", Src: "open", Dst: "closed"},
		},
	)
	fmt.Println(fsm.Is("closed"))
	fmt.Println(fsm.Is("open"))
	// Output:
	// true
	// false
}

func ExampleFSM_Can() {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"closed",
		EventTypeEvents{
			{Label: "open", Src: "closed", Dst: "open"},
			{Label: "close", Src: "open", Dst: "closed"},
		},
	)
	fmt.Println(fsm.Can("open"))
	fmt.Println(fsm.Can("close"))
	// Output:
	// true
	// false
}

//func ExampleFSM_AvailableTransitions() {
//	fsm := NewEventTypeStateTypeFiniteStateMachine(
//		"closed",
//		EventTypeEvents{
//			{Label: "open", Src: "closed", Dst: "open"},
//			{Label: "close", Src: "open", Dst: "closed"},
//			{Label: "kick", Src: "closed", Dst: "broken"},
//		},
//	)
//	// sort the results ordering is consistent for the output checker
//	transitions := fsm.AvailableTransitions()
//	sort.Strings(transitions)
//	fmt.Println(transitions)
//	// Output:
//	// [kick open]
//}

func ExampleFSM_Cannot() {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"closed",
		EventTypeEvents{
			{Label: "open", Src: "closed", Dst: "open"},
			{Label: "close", Src: "open", Dst: "closed"},
		},
	)
	fmt.Println(fsm.Cannot("open"))
	fmt.Println(fsm.Cannot("close"))
	// Output:
	// false
	// true
}

func ExampleFSM_Event() {
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"closed",
		EventTypeEvents{
			{Label: "open", Src: "closed", Dst: "open"},
			{Label: "close", Src: "open", Dst: "closed"},
		},
	)
	fmt.Println(fsm.Current())
	err := fsm.Event("open")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	err = fsm.Event("close")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	// Output:
	// closed
	// open
	// closed
}

func ExampleFSM_Transition() {
	leaveClosed := func(t Transition) error {
		t.SetAsync()
		return nil
	}
	fsm := NewEventTypeStateTypeFiniteStateMachine(
		"closed",
		EventTypeEvents{
			{Label: "open", Src: "closed", Dst: "open", LeaveState: leaveClosed},
			{Label: "close", Src: "open", Dst: "closed"},
		},
	)
	err := fsm.Event("open")
	if e, ok := err.(AsyncError); !ok && e.Err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	err = fsm.Transition()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	// Output:
	// closed
	// open
}
