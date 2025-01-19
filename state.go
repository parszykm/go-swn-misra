package main

import "sync"

type PingState string

const (
	PingLock PingState = "ping_lock"
	PingFree PingState = "ping_free"
)

type State struct {
	mu    sync.Mutex
	State PingState
}

func (state *State) ManageState(stateToSet PingState) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.State = stateToSet
}

func (state *State) GetState() PingState {
	state.mu.Lock()
	defer state.mu.Unlock()
	return state.State
}
