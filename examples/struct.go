// +build ignore

package main

import (
	"fmt"
	"github.com/looplab/fsm"
)

type Door struct {
	To  string
	FSM *fsm.EventTypeStateTypeFiniteStateMachine
}

func NewDoor(to string) *Door {
	d := &Door{
		To: to,
	}
	enterState := func(t fsm.Transition) error {
		d.enterState(t)
		return nil
	}
	d.FSM = fsm.NewEventTypeStateTypeFiniteStateMachine(
		"closed",
		fsm.EventTypeEvents{
			{Label: "open", Src: "closed", Dst: "open"},
			{Label: "close", Src: "open", Dst: "closed"},
		},
	)
	d.FSM.EnterState = enterState

	return d
}

func (d *Door) enterState(t fsm.Transition) {
	fmt.Printf("The door to %s is %s\n", d.To, t.Dst())
}

func main() {
	door := NewDoor("heaven")

	err := door.FSM.Event("open")
	if err != nil {
		fmt.Println(err)
	}

	err = door.FSM.Event("close")
	if err != nil {
		fmt.Println(err)
	}
}
