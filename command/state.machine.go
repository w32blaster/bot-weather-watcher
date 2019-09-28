package command

const (
	StepEnterLocation     = 1
	StepEnterMaxWindSpeed = 2
	StepEnterMinTemp      = 3
	FINISHED              = -1
)

type (
	state struct {
		next      int
		fnProcess func(string) string
	}

	StateMachine struct {
		UserID       string
		currentState int
	}
)

var states = map[int]state{

	StepEnterLocation: {
		next: StepEnterMaxWindSpeed,
		fnProcess: func(rawMessage string) string {
			// create a half-ready bookmark
			// if correct, then move to the next step
			return "Ok, now enter the max wind speed (m/s) that is comfortable for you in that location"
		},
	},

	StepEnterMaxWindSpeed: {
		next: StepEnterMinTemp,
		fnProcess: func(rawMessage string) string {
			return "Go it, now send me lowest temperature (in ˚C) that suits for you "
		},
	},

	StepEnterMinTemp: {
		next: FINISHED,
		fnProcess: func(rawMessage string) string {
			return "All done, this location was saved for you."
		},
	},
}

func LoadStateMachineFor(userID string) *StateMachine {

	return &StateMachine{
		UserID:       userID,
		currentState: loadState(userID), // load from Db
	}
}

// Move to the next state, updates internal state in case of success;
// returns message that should be returned to user
func (sm *StateMachine) NextState(rawMessage string) string {
	state := states[sm.currentState]
	return state.fnProcess(rawMessage)
}

func loadState(userID string) int {
	// load state from DB
	return -1
}
