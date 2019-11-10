// +build ignore

package main

import (
	"fmt"
	"github.com/looplab/fsm"
)

func main() {
	fsm := fsm.NewFSM(
		"idle",
		fsm.Events{
			{Label: "scan", Src: fsm.States{"idle"}, Dst: "scanning"},
			{Label: "working", Src: fsm.States{"scanning"}, Dst: "scanning"},
			{Label: "situation", Src: fsm.States{"scanning"}, Dst: "scanning"},
			{Label: "situation", Src: fsm.States{"idle"}, Dst: "idle"},
			{Label: "finish", Src: fsm.States{"scanning"}, Dst: "idle"},
		},
		fsm.Transitions{
			"scan": func(t fsm.Transition) error {
				//fmt.Println("after_scan: " + t.Current())
				return nil
			},
			"working": func(t fsm.Transition) error {
				//fmt.Println("working: " + t.Current())
				return nil
			},
			"situation": func(t fsm.Transition) error {
				//fmt.Println("situation: " + t.Current())
				return nil
			},
			"finish": func(t fsm.Transition) error {
				//fmt.Println("finish: " + t.Current())
				return nil
			},
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

	fmt.Printf("4:%s\n",fsm.Current())

}
