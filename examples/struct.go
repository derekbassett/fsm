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

	d.FSM = fsm.NewFSM(
		"closed",
		fsm.Events{
			{Label: "open", Src: fsm.States{"closed"}, Dst: "open"},
			{Label: "close", Src: fsm.States{"open"}, Dst: "closed"},
		},
		fsm.Transitions{
			"enter_state": func(t fsm.Transition) error {
				d.enterState(t)
				return nil
			},
		},
	)

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
