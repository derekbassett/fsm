// +build ignore

package main

import (
	"fmt"
	"github.com/looplab/fsm"
)

func main() {
	fsm := fsm.NewEventTypeStateTypeFiniteStateMachine(
		"closed",
		fsm.EventTypeEvents{
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
}
