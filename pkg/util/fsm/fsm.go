package fsm

import "fmt"

// internal state
// mapping of internal state X outgoing message -> internal state X outgoing message

type FSM struct {
	state         interface{}
	transitionMap map[TransitionInput]TransitionOutput
	recentMsg     string
	initState     interface{}
}

type TransitionInput struct {
	OldState interface{}
	InMsg    string
}

type TransitionOutput struct {
	NewState interface{}
	OutMsg   string
}

func NewFSM(diagram map[TransitionInput]TransitionOutput, initialState interface{}) *FSM {
	return &FSM{
		initialState,
		diagram,
		"",
		initialState,
	}
}

func (f *FSM) ApplyTransition(msg string) (string, bool) {
	out, ok := f.transitionMap[TransitionInput{
		f.state,
		msg,
	}]
	// if this row is not found in the state transition diagram, keep state
	if !ok {
		return f.recentMsg, false
	}
	f.state = out.NewState
	f.recentMsg = out.OutMsg
	return out.OutMsg, true
}

func (f *FSM) Reset() {
	f.state = f.initState
	f.recentMsg = ""
}

func (f *FSM) Log() {
	fmt.Printf("New state: %d, out msg: %s\n", f.state, f.recentMsg)
}
