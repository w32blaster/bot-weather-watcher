package command

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/w32blaster/bot-weather-watcher/structs"
	"log"
	"strconv"
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

			if _, err := strconv.Atoi(rawMessage); err != nil {
				return fmt.Sprintf("hey, %s is not a number! Please send me only number which is max speed of wind acceptable for you."+
					"Please ommit the 'm/s' or other suffixes", rawMessage)
			}

			sm.UpdateFieldInBookmark("LocationID", rawMessage)

			// if correct, then move to the next step
			if err := sm.markNextStepState(StepEnterMaxWindSpeed); err != nil {
				log.Println(err.Error())
				return "Internal error: can't update state"
			}

			fmt.Printf("%+v", sm.GetBookmark())
			return "Ok, now enter the max wind speed (m/s) that is comfortable for you in that location"
		},
	},

	StepEnterMaxWindSpeed: {
		next: StepEnterMinTemp,
		fnProcess: func(rawMessage string, sm *StateMachine) string {
			intMaxWindSpeed, err := strconv.Atoi(rawMessage)
			if err != nil {
				return fmt.Sprintf("hey, %s is not a number! Please send me only number which is max speed of wind acceptable for you."+
					"Please ommit the 'm/s' or other suffixes", rawMessage)
			}

			sm.UpdateFieldInBookmark("MaxWindSpeed", intMaxWindSpeed)
			sm.markNextStepState(StepEnterMinTemp)

			fmt.Printf("%+v", sm.GetBookmark())
			return "Go it, now send me lowest temperature (in ˚C) that suits for you "
		},
	},

	StepEnterMinTemp: {
		next: FINISHED,
		fnProcess: func(rawMessage string, sm *StateMachine) string {
			intMinTemp, err := strconv.Atoi(rawMessage)
			if err != nil {
				return fmt.Sprintf("hey, %s is not a number! Please send me only plain number which is min temperature acceptable for you."+
					"Please ommit any suffixes", rawMessage)
			}

			sm.UpdateFieldInBookmark("LowestTemp", intMinTemp)
			sm.UpdateFieldInBookmark("IsReady", true)

			sm.DeleteStateForCurrentUser()

			fmt.Printf("%+v", sm.GetBookmark())
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

func (sm *StateMachine) DeleteStateForCurrentUser() {
	state, err := sm.loadState(sm.UserID)
	if err != nil {
		log.Println(err.Error())
		return
	}
	sm.db.DeleteStruct(&state)
}

// Move to the next state, updates internal state in case of success;
// returns message that should be returned to user
func (sm *StateMachine) ProcessNextState(rawMessage string) string {
	state := states[sm.currentState]
	return state.fnProcess(rawMessage, sm)
}

func (sm *StateMachine) markNextStepState(newState int) error {

	// get current state
	var currState structs.UserState
	if err := sm.db.One("UserID", sm.UserID, &currState); err != nil {
		return err
	}

	// and update its value
	return sm.db.UpdateField(&currState, "CurrentState", newState)
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
