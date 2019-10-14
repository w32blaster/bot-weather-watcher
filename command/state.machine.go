package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/w32blaster/bot-weather-watcher/structs"

	"github.com/asdine/storm"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	StepEnterLocation     = 1
	StepEnterMaxWindSpeed = 2
	StepEnterMinTemp      = 3
	StepSpecifyDays       = 4
	FINISHED              = -1
	buttonOnlyWeekends    = "Only weekends"
	buttonAllDays         = "All days"
)

type (
	state struct {
		next      int
		fnProcess func(string, *StateMachine)
	}

	StateMachine struct {
		UserID       int
		currentState int
		db           *storm.DB
		bot          *tgbotapi.BotAPI
		chatID       int64
	}
)

var checkPeriods = map[string]int{
	buttonAllDays:      0,
	buttonOnlyWeekends: 1,
}

var states = map[int]state{

	StepEnterLocation: {
		next: StepEnterMaxWindSpeed,
		fnProcess: func(rawMessage string, sm *StateMachine) {

			if !strings.HasPrefix(rawMessage, LocationIDPrefix) {
				sendMsg(sm.bot, sm.chatID, "Error, wrong location ID format")
				return
			}

			locaIDClean := strings.TrimPrefix(rawMessage, LocationIDPrefix)
			if _, err := strconv.Atoi(locaIDClean); err != nil {
				sendMsg(sm.bot, sm.chatID, fmt.Sprintf("hey, %s is not a number! Please send me only number which is max speed of wind acceptable for you."+
					"Please ommit the 'mph' or other suffixes", rawMessage))
				return
			}

			if err := sm.UpdateFieldInBookmark("LocationID", locaIDClean); err != nil {
				log.WithError(err).Error("Can't update field for a given bookmark")
				sendMsg(sm.bot, sm.chatID, "Internal error: can't update location")
				return
			}

			// if correct, then move to the next step
			if err := sm.markNextStepState(StepEnterMaxWindSpeed); err != nil {
				log.WithError(err).Error("Can't update next step in state machine")
				sendMsg(sm.bot, sm.chatID, "Internal error: can't update state")
				return
			}

			sendMsg(sm.bot, sm.chatID, "Ok, now enter the max wind speed (mph) that is comfortable for you in that location")
		},
	},

	StepEnterMaxWindSpeed: {
		next: StepEnterMinTemp,
		fnProcess: func(rawMessage string, sm *StateMachine) {
			intMaxWindSpeed, err := strconv.Atoi(rawMessage)
			if err != nil {
				sendMsg(sm.bot, sm.chatID, fmt.Sprintf("hey, %s is not a number! Please send me only number which is min temperature acceptable for you."+
					"Please ommit the 'degrees, ˚C or other suffixes; number only", rawMessage))
				return
			}

			sm.UpdateFieldInBookmark("MaxWindSpeed", intMaxWindSpeed)
			sm.markNextStepState(StepEnterMinTemp)

			sendMsg(sm.bot, sm.chatID, "Go it, now send me lowest temperature (in ˚C) that suits for you ")
		},
	},

	StepEnterMinTemp: {
		next: StepSpecifyDays,
		fnProcess: func(rawMessage string, sm *StateMachine) {
			intMinTemp, err := strconv.Atoi(rawMessage)
			if err != nil {
				sendMsg(sm.bot, sm.chatID, fmt.Sprintf("hey, %s is not a number! Please send me only plain number which is min temperature acceptable for you."+
					"Please ommit any suffixes", rawMessage))
				return
			}

			sm.UpdateFieldInBookmark("LowestTemp", intMinTemp)
			sm.markNextStepState(StepSpecifyDays)

			msg := tgbotapi.NewMessage(sm.chatID, "Desired temperature is saved. The last step, what days do you want to observe? Only weekends (makes sense if you "+
				"on work during weekdays) or whole week (when you have a vacation or you have flexible time schedule)?")
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				[]tgbotapi.KeyboardButton{
					tgbotapi.NewKeyboardButton(buttonOnlyWeekends),
					tgbotapi.NewKeyboardButton(buttonAllDays),
				},
			)
			sm.bot.Send(msg)
		},
	},

	StepSpecifyDays: {
		next: FINISHED,
		fnProcess: func(rawMessage string, sm *StateMachine) {
			isValid := rawMessage == buttonOnlyWeekends || rawMessage == buttonAllDays
			if !isValid {
				sendMsg(sm.bot, sm.chatID, "Please click one of two buttons provided below")
				return
			}

			sm.UpdateFieldInBookmark("CheckPeriod", checkPeriods[rawMessage])
			sm.UpdateFieldInBookmark("IsReady", true)

			sm.currentState = FINISHED
			DeleteStateForUser(sm.db, sm.UserID)

			// send message and hide keyboard shown on the last step
			msg := tgbotapi.NewMessage(sm.chatID, "All done, this location was saved for you")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			sm.bot.Send(msg)
		},
	},
}

func LoadStateMachineFor(botApi *tgbotapi.BotAPI, chatID int64, userID int, stormDb *storm.DB) (*StateMachine, error) {

	sm := StateMachine{
		UserID: userID,
		db:     stormDb,
		bot:    botApi,
		chatID: chatID,
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
func (sm *StateMachine) ProcessNextState(rawMessage string) {
	state := states[sm.currentState]
	state.fnProcess(rawMessage, sm)
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
			log.WithError(err).
				WithField("user-id", userID).
				Error("attempt to create a new state and persist it to the database")
			return -1, err
		}
	}

	return state.CurrentState, nil
}
