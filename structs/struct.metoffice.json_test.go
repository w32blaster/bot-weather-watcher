package structs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFiveDayParsing(t *testing.T) {

	// Given
	confFile, _ := os.Open("../api-examples/example-5-day-forecast-aerodrome.json")
	bytes, err := ioutil.ReadAll(confFile)
	assert.Nil(t, err)

	// When
	var result RootSiteRep
	err = json.Unmarshal(bytes, &result)
	assert.NotNil(t, result)
	assert.Nil(t, err)
	//fmt.Println("ERROR: " + err.Error())

	// Then
	fmt.Printf("%+v", result)
}
