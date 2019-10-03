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
)

func TestStateMachineStepByStep(t *testing.T) {

	// Given:
	dir, db := prepareDB()
	defer os.RemoveAll(dir)
	defer db.Close()

	// When:
	sm, err := LoadStateMachineFor(UserID, db)
	assert.Nil(t, err)
	assert.NotNil(t, sm)

	// Step 1
	// When:
	err = sm.CreateNewBookmark()
	assert.Nil(t, err)

	// then:
	assert.Equal(t, StepEnterLocation, sm.currentState)

	// Step 2
	// When:
	msg := sm.ProcessNextState(LocationIDPrefix + TestLocationID)
	assert.NotEmpty(t, msg)

	// then:
	assert.Equal(t, StepEnterMaxWindSpeed, sm.currentState)

	// Step 3
	// When:
	msg = sm.ProcessNextState("20")
	assert.NotEmpty(t, msg)

	// then:
	assert.Equal(t, StepEnterMinTemp, sm.currentState)

	// Step 4
	// When:
	msg = sm.ProcessNextState("10")
	assert.NotEmpty(t, msg)

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
