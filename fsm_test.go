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
)

type fakeTransitionerObj struct {
}

func (t fakeTransitionerObj) Transition(f *EventTypeStateTypeFiniteStateMachine) error {
	return &InternalError{}
}

func TestSameState(t *testing.T) {
	start := new(StateType)
	run := new(EventType)
	fsm := NewFSM(
		start,
		Events{
			{Label: run, Src: States{start}, Dst: start},
		},
		Transitions{},
	)
	fsm.Event("run")
	if fsm.Current() != start {
		t.Error("expected state to be 'start'")
	}
}

func TestState(t *testing.T) {
	fsm := NewFSM(
		"walking",
		Events{
			{Label: "walk", Src: States{"start"}, Dst: "walking"},
		},
		Transitions{},
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
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "running"},
		},
		Transitions{},
	)
	fsm.transitionerObj = new(fakeTransitionerObj)
	err := fsm.Event("run")
	if err == nil {
		t.Error("bad transition should give an error")
	}
}

func TestInappropriateEvent(t *testing.T) {
	fsm := NewFSM(
		"closed",
		Events{
			{Label: "open", Src: States{"closed"}, Dst: "open"},
			{Label: "close", Src: States{"open"}, Dst: "closed"},
		},
		Transitions{},
	)
	err := fsm.Event("close")
	if e, ok := err.(InvalidEventError); !ok && e.Event != "close" && e.State != "closed" {
		t.Error("expected 'InvalidEventError' with correct state and event")
	}
}

func TestInvalidEvent(t *testing.T) {
	fsm := NewFSM(
		"closed",
		Events{
			{Label: "open", Src: States{"closed"}, Dst: "open"},
			{Label: "close", Src: States{"open"}, Dst: "closed"},
		},
		Transitions{},
	)
	err := fsm.Event("lock")
	if e, ok := err.(UnknownEventError); !ok && e.Event != "close" {
		t.Error("expected 'UnknownEventError' with correct event")
	}
}

func TestMultipleSources(t *testing.T) {
	fsm := NewFSM(
		"one",
		Events{
			{Label: "first", Src: States{"one"}, Dst: "two"},
			{Label: "second", Src: States{"two"}, Dst: "three"},
			{Label: "reset", Src: States{"one", "two", "three"}, Dst: "one"},
		},
		Transitions{},
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
	fsm := NewFSM(
		"start",
		Events{
			{Label: "first", Src: States{"start"}, Dst: "one"},
			{Label: "second", Src: States{"start"}, Dst: "two"},
			{Label: "reset", Src: States{"one"}, Dst: "reset_one"},
			{Label: "reset", Src: States{"two"}, Dst: "reset_two"},
			{Label: "reset", Src: States{"reset_one", "reset_two"}, Dst: "start"},
		},
		Transitions{},
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

	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"before_event": func(t Transition) error {
				beforeEvent = true
				return nil
			},
			"leave_state": func(t Transition) error {
				leaveState = true
				return nil
			},
			"enter_state": func(t Transition) error {
				enterState = true
				return nil
			},
			"after_event": func(t Transition) error {
				afterEvent = true
				return nil
			},
		},
	)

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

	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"before_run": func(t Transition) error {
				beforeEvent = true
				return nil
			},
			"leave_start": func(t Transition) error {
				leaveState = true
				return nil
			},
			"enter_end": func(t Transition) error {
				enterState = true
				return nil
			},
			"after_run": func(t Transition) error {
				afterEvent = true
				return nil
			},
		},
	)

	fsm.Event("run")
	if !(beforeEvent && leaveState && enterState && afterEvent) {
		t.Error("expected all callbacks to be called")
	}
}

func TestSpecificCallbacksShortform(t *testing.T) {
	enterState := false
	afterEvent := false

	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"end": func(t Transition) error {
				enterState = true
				return nil
			},
			"run": func(t Transition) error {
				afterEvent = true
				return nil
			},
		},
	)

	fsm.Event("run")
	if !(enterState && afterEvent) {
		t.Error("expected all callbacks to be called")
	}
}

func TestBeforeEventWithoutTransition(t *testing.T) {
	beforeEvent := true

	fsm := NewFSM(
		"start",
		Events{
			{Label: "dontrun", Src: States{"start"}, Dst: "start"},
		},
		Transitions{
			"before_event": func(t Transition) error {
				beforeEvent = true
				return nil
			},
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
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"before_event": func(t Transition) error {
				t.Cancel()
				return nil
			},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestCancelBeforeSpecificEvent(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"before_run": func(t Transition) error {
				t.Cancel()
				return nil
			},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestCancelLeaveGenericState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"leave_state": func(t Transition) error {
				t.Cancel()
				return nil
			},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestCancelLeaveSpecificState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"leave_start": func(t Transition) error {
				t.Cancel()
				return nil
			},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestCancelWithError(t *testing.T) {
	expect := fmt.Errorf("error")
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"before_event": func(t Transition) error {
				t.Cancel()
				return expect
			},
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
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"leave_state": func(t Transition) error {
				t.SetAsync()
				return nil
			},
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

func TestAsyncTransitionSpecificState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"leave_start": func(t Transition) error {
				t.SetAsync()
				return nil
			},
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
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
			{Label: "reset", Src: States{"end"}, Dst: "start"},
		},
		Transitions{
			"leave_start": func(t Transition) error {
				t.SetAsync()
				return nil
			},
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
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
			{Label: "reset", Src: States{"end"}, Dst: "start"},
		},
		Transitions{},
	)
	err := fsm.Transition()
	if _, ok := err.(NotInTransitionError); !ok {
		t.Error("expected 'NotInTransitionError'")
	}
}

func TestCallbackNoError(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"run": func(t Transition) error {
				return nil
			},
		},
	)
	e := fsm.Event("run")
	if e != nil {
		t.Error("expected no error")
	}
}

func TestCallbackError(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"run": func(t Transition) error {
				return fmt.Errorf("error")
			},
		},
	)
	e := fsm.Event("run")
	if e == nil || e.Error() != "error" {
		t.Error("expected error to be 'error'")
	}
}

//func TestCallbackArgs(t *testing.T) {
//	fsm := NewFSM(
//		"start",
//		Events{
//			{Label: "run", Src: States{"start"}, Dst: "end"},
//		},
//		Transitions{
//			"run": func(t Transition) {
//				if len(e.Args) != 1 {
//					t.Error("too few arguments")
//				}
//				arg, ok := e.Args[0].(string)
//				if !ok {
//					t.Error("not a string argument")
//				}
//				if arg != "test" {
//					t.Error("incorrect argument")
//				}
//			},
//		},
//	)
//	fsm.Event("run", "test")
//}

func TestNoDeadLock(t *testing.T) {
	var fsm *EventTypeStateTypeFiniteStateMachine
	fsm = NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"run": func(t Transition) error {
				fsm.Current() // Should not result in a panic / deadlock
				return nil
			},
		},
	)
	fsm.Event("run")
}

func TestThreadSafetyRaceCondition(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "end"},
		},
		Transitions{
			"run": func(t Transition) error {
				return nil
			},
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

//func TestDoubleTransition(t *testing.T) {
//	var fsm *EventTypeStateTypeFiniteStateMachine
//	var wg sync.WaitGroup
//	wg.Add(2)
//	fsm = NewFSM(
//		"start",
//		Events{
//			{Label: "run", Src: States{"start"}, Dst: "end"},
//		},
//		Transitions{
//			"before_run": func(t Transition) error {
//				wg.Done()
//				// Imagine a concurrent event coming in of the same type while
//				// the data access mutex is unlocked because the current transition
//				// is running its event callbacks, getting around the "active"
//				// transition checks
//				if len(e.Args) == 0 {
//					// Must be concurrent so the test may pass when we add a mutex that synchronizes
//					// calls to Transition(...). It will then fail as an inappropriate transition as we
//					// have changed state.
//					go func() {
//						if err := fsm.Event("run", "second run"); err != nil {
//							fmt.Println(err)
//							wg.Done() // It should fail, and then we unfreeze the test.
//						}
//					}()
//					time.Sleep(20 * time.Millisecond)
//				} else {
//					panic("Was able to reissue an event mid-transition")
//				}
//			},
//		},
//	)
//	if err := fsm.Event("run"); err != nil {
//		fmt.Println(err)
//	}
//	wg.Wait()
//}

func TestNoTransition(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Label: "run", Src: States{"start"}, Dst: "start"},
		},
		Transitions{},
	)
	err := fsm.Event("run")
	if _, ok := err.(NoTransitionError); !ok {
		t.Error("expected 'NoTransitionError'")
	}
}

func ExampleNewFSM() {
	fsm := NewFSM(
		"green",
		Events{
			{Label: "warn", Src: States{"green"}, Dst: "yellow"},
			{Label: "panic", Src: States{"yellow"}, Dst: "red"},
			{Label: "panic", Src: States{"green"}, Dst: "red"},
			{Label: "calm", Src: States{"red"}, Dst: "yellow"},
			{Label: "clear", Src: States{"yellow"}, Dst: "green"},
		},
		Transitions{
			"before_warn": func(t Transition) error {
				fmt.Println("before_warn")
				return nil
			},
			"before_event": func(t Transition) error {
				fmt.Println("before_event")
				return nil
			},
			"leave_green": func(t Transition) error {
				fmt.Println("leave_green")
				return nil
			},
			"leave_state": func(t Transition) error {
				fmt.Println("leave_state")
				return nil
			},
			"enter_yellow": func(t Transition) error {
				fmt.Println("enter_yellow")
				return nil
			},
			"enter_state": func(t Transition) error {
				fmt.Println("enter_state")
				return nil
			},
			"after_warn": func(t Transition) error {
				fmt.Println("after_warn")
				return nil
			},
			"after_event": func(t Transition) error {
				fmt.Println("after_event")
				return nil
			},
		},
	)
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
	fsm := NewFSM(
		"closed",
		Events{
			{Label: "open", Src: States{"closed"}, Dst: "open"},
			{Label: "close", Src: States{"open"}, Dst: "closed"},
		},
		Transitions{},
	)
	fmt.Println(fsm.Current())
	// Output: closed
}

func ExampleFSM_Is() {
	fsm := NewFSM(
		"closed",
		Events{
			{Label: "open", Src: States{"closed"}, Dst: "open"},
			{Label: "close", Src: States{"open"}, Dst: "closed"},
		},
		Transitions{},
	)
	fmt.Println(fsm.Is("closed"))
	fmt.Println(fsm.Is("open"))
	// Output:
	// true
	// false
}

func ExampleFSM_Can() {
	fsm := NewFSM(
		"closed",
		Events{
			{Label: "open", Src: States{"closed"}, Dst: "open"},
			{Label: "close", Src: States{"open"}, Dst: "closed"},
		},
		Transitions{},
	)
	fmt.Println(fsm.Can("open"))
	fmt.Println(fsm.Can("close"))
	// Output:
	// true
	// false
}

//func ExampleFSM_AvailableTransitions() {
//	fsm := NewFSM(
//		"closed",
//		Events{
//			{Label: "open", Src: States{"closed"}, Dst: "open"},
//			{Label: "close", Src: States{"open"}, Dst: "closed"},
//			{Label: "kick", Src: States{"closed"}, Dst: "broken"},
//		},
//		Transitions{},
//	)
//	// sort the results ordering is consistent for the output checker
//	transitions := fsm.AvailableTransitions()
//	sort.Strings(transitions)
//	fmt.Println(transitions)
//	// Output:
//	// [kick open]
//}

func ExampleFSM_Cannot() {
	fsm := NewFSM(
		"closed",
		Events{
			{Label: "open", Src: States{"closed"}, Dst: "open"},
			{Label: "close", Src: States{"open"}, Dst: "closed"},
		},
		Transitions{},
	)
	fmt.Println(fsm.Cannot("open"))
	fmt.Println(fsm.Cannot("close"))
	// Output:
	// false
	// true
}

func ExampleFSM_Event() {
	fsm := NewFSM(
		"closed",
		Events{
			{Label: "open", Src: States{"closed"}, Dst: "open"},
			{Label: "close", Src: States{"open"}, Dst: "closed"},
		},
		Transitions{},
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
	fsm := NewFSM(
		"closed",
		Events{
			{Label: "open", Src: States{"closed"}, Dst: "open"},
			{Label: "close", Src: States{"open"}, Dst: "closed"},
		},
		Transitions{
			"leave_closed": func(t Transition) error {
				t.SetAsync()
				return nil
			},
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
