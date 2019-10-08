package command

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/w32blaster/bot-weather-watcher/structs"
)

func (sm *StateMachine) CreateNewBookmark(chatID int64) error {

	if bookmark := sm.GetUnfinishedBookmark(); bookmark != nil {
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
		ChatID:       chatID,
	})
}

func (sm *StateMachine) UpdateFieldInBookmark(fieldName string, value interface{}) error {
	bookmark := sm.GetUnfinishedBookmark()
	return sm.db.UpdateField(bookmark, fieldName, value)
}

func (sm *StateMachine) GetUnfinishedBookmark() *structs.UsersLocationBookmark {
	var bookmark structs.UsersLocationBookmark
	err := sm.db.Select(q.And(
		q.Eq("UserID", sm.UserID),
		q.Eq("IsReady", false),
	)).First(&bookmark)

	if err != nil {
		return nil
	}
	return &bookmark
}

func DeleteStateForUser(db *storm.DB, userID int) {
	var state structs.UserState
	if err := db.One("UserID", userID, &state); err == nil {
		db.DeleteStruct(&state)
	}
}

func DeleteAllUnfinishedBookmarksForThisUser(db *storm.DB, userID int) {
	query := db.Select(q.Eq("UserID", userID), q.Eq("IsReady", false))
	query.Delete(new(structs.UsersLocationBookmark))
}
