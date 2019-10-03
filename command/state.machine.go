package command

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/w32blaster/bot-weather-watcher/structs"
	"log"
	"strconv"
	"strings"
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

			if !strings.HasPrefix(rawMessage, LocationIDPrefix) {
				return "Error, wrong location ID format"
			}

			locaIDClean := strings.TrimPrefix(rawMessage, LocationIDPrefix)
			if _, err := strconv.Atoi(locaIDClean); err != nil {
				return fmt.Sprintf("hey, %s is not a number! Please send me only number which is max speed of wind acceptable for you."+
					"Please ommit the 'mph' or other suffixes", rawMessage)
			}

			if err := sm.UpdateFieldInBookmark("LocationID", locaIDClean); err != nil {
				log.Println(err.Error())
				return "Internal error: can't update location"
			}

			// if correct, then move to the next step
			if err := sm.markNextStepState(StepEnterMaxWindSpeed); err != nil {
				log.Println(err.Error())
				return "Internal error: can't update state"
			}

			fmt.Printf("%+v", sm.GetUnfinishedBookmark())
			return "Ok, now enter the max wind speed (mph) that is comfortable for you in that location"
		},
	},

	StepEnterMaxWindSpeed: {
		next: StepEnterMinTemp,
		fnProcess: func(rawMessage string, sm *StateMachine) string {
			intMaxWindSpeed, err := strconv.Atoi(rawMessage)
			if err != nil {
				return fmt.Sprintf("hey, %s is not a number! Please send me only number which is min temperature acceptable for you."+
					"Please ommit the 'degrees, ˚C or other suffixes; number only", rawMessage)
			}

			sm.UpdateFieldInBookmark("MaxWindSpeed", intMaxWindSpeed)
			sm.markNextStepState(StepEnterMinTemp)

			fmt.Printf("%+v", sm.GetUnfinishedBookmark())
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

			sm.currentState = FINISHED
			DeleteStateForUser(sm.db, sm.UserID)

			fmt.Printf("%+v", sm.GetUnfinishedBookmark())
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

func (sm *StateMachine) markNextStepState(newState int) error {

	// get current state
	var currState structs.UserState
	if err := sm.db.One("UserID", sm.UserID, &currState); err != nil {
		return err
	}

	// and update its value
	if err := sm.db.UpdateField(&currState, "CurrentState", newState); err != nil {
		return err
	}

	sm.currentState = newState
	return nil
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
