package command

import (
	"github.com/w32blaster/bot-weather-watcher/structs"
)

func (sm *StateMachine) CreateNewBookmark() error {
	var bookmark structs.UsersLocationBookmark
	if err := sm.db.One("UserID", sm.UserID, &bookmark); err == nil {
		return nil // assuming, this is not an error, just object already created and no actions are needed
	}

	return sm.db.Save(&structs.UsersLocationBookmark{UserID: sm.UserID})
}

func (sm *StateMachine) UpdateFieldInBookmark(fieldName string, value interface{}) error {
	return sm.db.UpdateField(&structs.UsersLocationBookmark{UserID: sm.UserID}, "Age", 0)
}
