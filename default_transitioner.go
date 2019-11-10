package fsm


// defaultTransitioner is the default implementation of the transitioner
// interface. Other implementations can be swapped in for testing.
type defaultTransitioner struct{}

var _ EventTypeEventStateTypeStateTransitioner = (*defaultTransitioner)(nil)

// Transition completes an asynchrounous state change.
//
// The callback for leave_<STATE> must previously have called Async on its
// event to have initiated an asynchronous state transition.
func (t defaultTransitioner) Transition(f *EventTypeStateTypeFiniteStateMachine) error {
	if f.transition == nil {
		return NotInTransitionError{}
	}
	f.transition()
	f.transition = nil
	return nil
}
