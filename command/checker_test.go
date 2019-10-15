package command

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWeekdayBother(t *testing.T) {

	var dataSet = []struct {
		dayChoice      int
		weekday        time.Weekday
		expectedResult bool
	}{
		{onlyWeekends, time.Monday, false},
		{onlyWeekends, time.Tuesday, false},
		{onlyWeekends, time.Wednesday, false},
		{onlyWeekends, time.Thursday, false},
		{onlyWeekends, time.Friday, true}, // because Friday is almost not working day in fact :)
		{onlyWeekends, time.Saturday, true},
		{onlyWeekends, time.Sunday, true},

		{allDays, time.Monday, true},
		{allDays, time.Tuesday, true},
		{allDays, time.Wednesday, true},
		{allDays, time.Thursday, true},
		{allDays, time.Friday, true},
		{allDays, time.Saturday, true},
		{allDays, time.Sunday, true},
	}

	for _, tt := range dataSet {
		t.Run(fmt.Sprintf("Should we bother if choise is %d at %s", tt.dayChoice, tt.weekday.String()),
			func(t *testing.T) {

				// When:
				result := shouldBotherForWeekdays(tt.dayChoice, tt.weekday)

				// Then:
				assert.Equal(t, tt.expectedResult, result)
			})
	}

}
