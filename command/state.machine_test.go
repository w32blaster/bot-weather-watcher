package command

import (
	"github.com/asdine/storm"
	"github.com/stretchr/testify/assert"
	"github.com/w32blaster/bot-weather-watcher/structs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const (
	TestLocationID = "111"
	UserID         = 111
	User2ID        = 222
)

func TestStateMachineStepByStep(t *testing.T) {

	// Given:
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	// When:
	sm, err := LoadStateMachineFor(nil, 1, UserID, db)
	assert.Nil(t, err)
	assert.NotNil(t, sm)

	// Step 1
	// When:
	err = sm.CreateNewBookmark(-1)
	assert.Nil(t, err)

	// then:
	assert.Equal(t, StepEnterLocation, sm.currentState)

	// Step 2
	// When:
	sm.ProcessNextState(LocationIDPrefix + TestLocationID)

	// then:
	assert.Equal(t, StepEnterMaxWindSpeed, sm.currentState)

	// Step 3
	// When:
	sm.ProcessNextState("20")

	// then:
	assert.Equal(t, StepEnterMinTemp, sm.currentState)

	// Step 4
	// When:
	sm.ProcessNextState("10")

	// then:
	assert.Equal(t, FINISHED, sm.currentState)

	// make sure we have only one saved (bookmarked) location in the database
	var bookmarks []structs.UsersLocationBookmark
	assert.Nil(t, db.All(&bookmarks))
	assert.Equal(t, 1, len(bookmarks))

	// and the bookmark contains valid information
	assert.Equal(t, TestLocationID, bookmarks[0].LocationID)
	assert.Equal(t, UserID, bookmarks[0].UserID)
	assert.Equal(t, 20, bookmarks[0].MaxWindSpeed)
	assert.Equal(t, 10, bookmarks[0].LowestTemp)
	assert.True(t, bookmarks[0].IsReady)
}

// scenario: two users, two bookmarks
func TestStateMachineForTwoUsers(t *testing.T) {

	// Given:
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	// State machine flow for user 1
	sm, _ := LoadStateMachineFor(nil, 0, UserID, db)
	sm.CreateNewBookmark(-1)
	assert.Equal(t, StepEnterLocation, sm.currentState)

	sm.ProcessNextState(LocationIDPrefix + TestLocationID)
	assert.Equal(t, StepEnterMaxWindSpeed, sm.currentState)

	sm.ProcessNextState("20")
	assert.Equal(t, StepEnterMinTemp, sm.currentState)

	sm.ProcessNextState("10")
	assert.Equal(t, FINISHED, sm.currentState)

	// Repeat the same for User 2
	sm, _ = LoadStateMachineFor(nil, 0, User2ID, db)
	sm.CreateNewBookmark(-1)
	assert.Equal(t, StepEnterLocation, sm.currentState)

	sm.ProcessNextState(LocationIDPrefix + TestLocationID)
	assert.Equal(t, StepEnterMaxWindSpeed, sm.currentState)

	sm.ProcessNextState("30")
	assert.Equal(t, StepEnterMinTemp, sm.currentState)

	sm.ProcessNextState("5")
	assert.Equal(t, FINISHED, sm.currentState)

	// make sure we have two bookmarks in the database
	var bookmarks []structs.UsersLocationBookmark
	assert.Nil(t, db.All(&bookmarks))
	assert.Equal(t, 2, len(bookmarks))

	// and the bookmark contains valid information
	assert.Equal(t, TestLocationID, bookmarks[0].LocationID)
	assert.Equal(t, UserID, bookmarks[0].UserID)
	assert.Equal(t, 20, bookmarks[0].MaxWindSpeed)
	assert.Equal(t, 10, bookmarks[0].LowestTemp)
	assert.True(t, bookmarks[0].IsReady)

	// and the bookmark contains valid information
	assert.Equal(t, TestLocationID, bookmarks[1].LocationID)
	assert.Equal(t, User2ID, bookmarks[1].UserID)
	assert.Equal(t, 30, bookmarks[1].MaxWindSpeed)
	assert.Equal(t, 5, bookmarks[1].LowestTemp)
	assert.True(t, bookmarks[1].IsReady)
}

// scenario: two users, two bookmarks
func TestStateMachineForTwoUsersNotFinished(t *testing.T) {

	// Given:
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	// State machine flow for user 1
	sm, _ := LoadStateMachineFor(nil, 1, UserID, db)
	sm.CreateNewBookmark(-1)
	assert.Equal(t, StepEnterLocation, sm.currentState)

	sm.ProcessNextState(LocationIDPrefix + TestLocationID)
	assert.Equal(t, StepEnterMaxWindSpeed, sm.currentState)

	sm.ProcessNextState("20")
	assert.Equal(t, StepEnterMinTemp, sm.currentState)

	sm.ProcessNextState("10")
	assert.Equal(t, FINISHED, sm.currentState)

	// Repeat the same for User 2
	sm, _ = LoadStateMachineFor(nil, 1, User2ID, db)
	sm.CreateNewBookmark(-1)
	assert.Equal(t, StepEnterLocation, sm.currentState)

	sm.ProcessNextState(LocationIDPrefix + TestLocationID)
	assert.Equal(t, StepEnterMaxWindSpeed, sm.currentState)

	sm.ProcessNextState("30")
	assert.Equal(t, StepEnterMinTemp, sm.currentState)

	// make sure we have two bookmarks in the database
	var bookmarks []structs.UsersLocationBookmark
	assert.Nil(t, db.All(&bookmarks))
	assert.Equal(t, 2, len(bookmarks))

	// but the first bookmark is ready
	assert.Equal(t, TestLocationID, bookmarks[0].LocationID)
	assert.Equal(t, UserID, bookmarks[0].UserID)
	assert.Equal(t, 20, bookmarks[0].MaxWindSpeed)
	assert.Equal(t, 10, bookmarks[0].LowestTemp)
	assert.True(t, bookmarks[0].IsReady)

	// ...and the second is not
	assert.Equal(t, TestLocationID, bookmarks[1].LocationID)
	assert.Equal(t, User2ID, bookmarks[1].UserID)
	assert.Equal(t, 30, bookmarks[1].MaxWindSpeed)
	assert.Equal(t, 0, bookmarks[1].LowestTemp) // default value, not filled yet
	assert.False(t, bookmarks[1].IsReady)       // <- False!
}

func prepareDB() (string, *storm.DB) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))

	db.Save(&structs.SiteLocation{
		ID:           TestLocationID,
		Elevation:    "10.0",
		Latitude:     "10.0",
		Longitude:    "20.0",
		Name:         "London",
		Region:       "SW1",
		AuthArea:     "Some testing area",
		NationalPark: "",
	})

	return dir, db
}
