package fsm



type transitionType int
const (
	callbackNone transitionType = iota
	callbackBeforeEvent
	callbackLeaveState
	callbackEnterState
	callbackAfterEvent
)

// cKey is a struct key used for keeping the callbacks mapped to a target.
type cKey struct {
	// target is either the name of a state or an event depending on which
	// callback type the key refers to. It can also be "" for a non-targeted
	// callback like before_event.
	target interface{}

	// callbackType is the situation when the callback will be run.
	callbackType transitionType
}

// eKey is a struct key used for storing the transition map.
type eKey struct {
	// event is the name of the event that the keys refers to.
	event EventType

	// src is the source from where the event can transition.
	src StateType
}