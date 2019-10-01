package command

import (
	"github.com/w32blaster/bot-weather-watcher/structs"
)

func (sm *StateMachine) CreateNewBookmark() error {

	if bookmark := sm.GetBookmark(); bookmark != nil {
		// the object already exists. Probably, user enters something wrong and decided to start again.
		// Remove old object to begin from the start with fresh state
		sm.db.DeleteStruct(&bookmark)
	}

	return sm.db.Save(&structs.UsersLocationBookmark{
		UserID:       sm.UserID,
		LocationID:   "",
		LowestTemp:   0,
		MaxWindSpeed: 0,
		IsReady:      false,
	})
}

func (sm *StateMachine) UpdateFieldInBookmark(fieldName string, value interface{}) error {
	bookmark := sm.GetBookmark()
	return sm.db.UpdateField(bookmark, fieldName, value)
}

func (sm *StateMachine) GetBookmark() *structs.UsersLocationBookmark {
	var bookmark structs.UsersLocationBookmark
	if err := sm.db.One("UserID", sm.UserID, &bookmark); err != nil {
		return nil
	}
	return &bookmark
}
