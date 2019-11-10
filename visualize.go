package fsm

import (
	"fmt"
	"io"
)

// Visualize outputs a visualization of a EventTypeStateTypeFiniteStateMachine in Graphviz format.
func (f *EventTypeStateTypeFiniteStateMachine) Visualize(w io.Writer) {

	states := make(map[StateType]int)

	w.Write([]byte(fmt.Sprintf(`digraph fsm {`)))
	w.Write([]byte("\n"))

	// make sure the initial state is at top
	for k, v := range f.transitions {
		if k.src == f.current {
			states[k.src]++
			states[v]++
			w.Write([]byte(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, k.src, v, k.event)))
			w.Write([]byte("\n"))
		}
	}

	for k, v := range f.transitions {
		if k.src != f.current {
			states[k.src]++
			states[v]++
			w.Write([]byte(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, k.src, v, k.event)))
			w.Write([]byte("\n"))
		}
	}

	w.Write([]byte("\n"))

	for k := range states {
		w.Write([]byte(fmt.Sprintf(`    "%s";`, k)))
		w.Write([]byte("\n"))
	}
	w.Write([]byte(fmt.Sprintln("}")))
}
