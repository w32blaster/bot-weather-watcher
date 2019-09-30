package command

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/w32blaster/bot-weather-watcher/structs"
	"log"
)

const (
	StepEnterLocation     = 1
	StepEnterMaxWindSpeed = 2
	StepEnterMinTemp      = 3
	FINISHED              = -1
)

type (
	state struct {
		next      int
		fnProcess func(string, *StateMachine) string
	}

	StateMachine struct {
		UserID       int
		currentState int
		db           *storm.DB
	}
)

var states = map[int]state{

	StepEnterLocation: {
		next: StepEnterMaxWindSpeed,
		fnProcess: func(rawMessage string, sm *StateMachine) string {

			// create a half-ready bookmark
			if err := sm.CreateNewBookmark(); err != nil {
				log.Println(err.Error())
				return "Whoops, can't create this bookmark for you, sorry. Try again?"
			}

			// if correct, then move to the next step
			sm.markNextStepState(StepEnterMaxWindSpeed, sm.UserID)
			return "Ok, now enter the max wind speed (m/s) that is comfortable for you in that location"
		},
	},

	StepEnterMaxWindSpeed: {
		next: StepEnterMinTemp,
		fnProcess: func(rawMessage string, sm *StateMachine) string {

			return "Go it, now send me lowest temperature (in ˚C) that suits for you "
		},
	},

	StepEnterMinTemp: {
		next: FINISHED,
		fnProcess: func(rawMessage string, sm *StateMachine) string {
			return "All done, this location was saved for you."
		},
	},
}

func LoadStateMachineFor(userID int, stormDb *storm.DB) (*StateMachine, error) {

	sm := StateMachine{
		UserID: userID,
		db:     stormDb,
	}

	currState, err := sm.loadState(userID)
	if err != nil {
		return nil, err
	}

	sm.currentState = currState
	return &sm, nil
}

// Move to the next state, updates internal state in case of success;
// returns message that should be returned to user
func (sm *StateMachine) ProcessNextState(rawMessage string) string {
	state := states[sm.currentState]
	return state.fnProcess(rawMessage, sm)
}

func (sm *StateMachine) markNextStepState(state, userID int) error {
	return sm.db.UpdateField(&structs.UserState{UserID: userID}, "CurrentState", state)
}

func (sm *StateMachine) loadState(userID int) (int, error) {

	// load state from DB
	var state structs.UserState
	if err := sm.db.One("UserID", userID, &state); err != nil {

		// existing state is not found, create a new one
		state = structs.UserState{
			UserID:       userID,
			CurrentState: StepEnterLocation,
		}
		if err := sm.db.Save(&state); err != nil {
			fmt.Printf("attempt to create a new state and persist in the database, but error occurred: %s \n", err.Error())
			return -1, err
		}
	}

	return state.CurrentState, nil
}
