package structs

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	//"github.com/stretchr/testify/assert"
)

func TestFiveDayParsing(t *testing.T) {

	// Given
	confFile, _ := os.Open("../api-examples/example-5-day-forecast-aerodrome.json")

	// When
	var result RootSiteRep
	json.NewDecoder(confFile).Decode(&result)

	// Then
	fmt.Printf("%+v", result)
	t.Fail()

	// ne rabotaet: slices, float and maps
}
