package main

type StateType string

const (
	CREATE_EVENT    StateType = "CREATE_EVENT"
	UPDATE_EVENT    StateType = "UPDATE_EVENT"
	ADD_ACTIVITY    StateType = "ADD_ACTIVITY"
	UPDATE_ACTIVITY StateType = "UPDATE_ACTIVITY"
	DELETE_ACTIVITY StateType = "DELETE_ACTIVITY"
)

// State tracking for each user
type UserState struct {
	Step      int
	StateType StateType
	Event     Event
	Activity  Activity
}

// Map to track state for each user
var userStates = make(map[string]*UserState)
