// +build ignore

package main

import (
	"fmt"
	"github.com/looplab/fsm"
)

func main() {
	scan := func(t fsm.Transition) error {
		// fmt.Println("after_scan: " + t.Current())
		return nil
	}
	working := func(t fsm.Transition) error {
		//fmt.Println("working: " + t.Current())
		return nil
	}
	situation := func(t fsm.Transition) error {
		//fmt.Println("situation: " + t.Current())
		return nil
	}
	finish := func(t fsm.Transition) error {
		//fmt.Println("finish: " + t.Current())
		return nil
	}

	fsm := fsm.NewEventTypeStateTypeFiniteStateMachine(
		"idle",
		fsm.EventTypeEvents{
			{Label: "scan", Src: "idle", Dst: "scanning", AfterEvent: scan},
			{Label: "working", Src: "scanning", Dst: "scanning", AfterEvent:working},
			{Label: "situation", Src: "scanning", Dst: "scanning", AfterEvent: situation},
			{Label: "situation", Src: "idle", Dst: "idle", AfterEvent: situation},
			{Label: "finish", Src: "scanning", Dst: "idle", AfterEvent: finish},
		},
	)

	fmt.Println(fsm.Current())

	err := fsm.Event("scan")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("1:%s\n", fsm.Current())

	err = fsm.Event("working")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("2:%s\n", fsm.Current())

	err = fsm.Event("situation")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("3:%s\n", fsm.Current())

	err = fsm.Event("finish")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("4:%s\n", fsm.Current())

}
