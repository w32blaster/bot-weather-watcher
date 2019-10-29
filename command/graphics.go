package command

import (
	"bytes"
	"github.com/guptarohit/asciigraph"
	"github.com/w32blaster/bot-weather-watcher/structs"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	vertTopLine         = "╭─────┬───────────────────"
	layoutMetofficeDate = "2006-01-02Z"
)

type weatherType struct {
	name string
	icon rune
}

// Please refer to the documentation for the code list
var mapWeatherTypes = map[int]weatherType{
	0:  {"Clear night", '🌖'},
	1:  {"Sunny day", '☀'},
	2:  {"Partly cloudy (night)", '🌤'},
	3:  {"Partly cloudy (day)", '🌤'},
	4:  {"Not used", '-'},
	5:  {"Mist", '🌫'},
	6:  {"Fog", '🌫'},
	7:  {"Cloudy", '⛅'},
	8:  {"Overcast", '☁'},
	9:  {"Light rain shower (night)", '🌧'},
	10: {"Light rain shower (day)", '🌧'},
	11: {"Drizzle", '🌧'},
	12: {"Light rain", '🌧'},
	13: {"Heavy rain shower (night)", '🌧'},
	14: {"Heavy rain shower (day)", '🌧'},
	15: {"Heavy rain", '🌧'},
	16: {"Sleet shower (night)", '🌨'},
	17: {"Sleet shower (day)", '🌨'},
	18: {"Sleet", '🌨'},
	19: {"Hail shower (night)", '🌨'},
	20: {"Hail shower (day)", '🌨'},
	21: {"Hail", '🌨'},
	22: {"Light snow shower (night)", '❄'},
	23: {"Light snow shower (day)", '❄'},
	24: {"Light snow", '❄'},
	25: {"Heavy snow shower (night)", '❄'},
	26: {"Heavy snow shower (day)", '❄'},
	27: {"Heavy snow", '❄'},
	28: {"Thunder shower (night)", '⛈'},
	29: {"Thunder shower (day)", '⛈'},
	30: {"Thunder", '🌩'},
}

func drawFiveDaysTable(root *structs.RootSiteRep) string {
	days := root.SiteRep.Dv.Location.Periods
	if len(days) != 5 {
		return ""
	}

	var buffer bytes.Buffer
	buffer.WriteString("```\n╭─────┬────────────────────╮ \n")

	// row 1 and 2
	var bufferRow1 bytes.Buffer
	var bufferRow2 bytes.Buffer
	var bufferRow3 bytes.Buffer
	var bufferRow4 bytes.Buffer

	for i, day := range days {

		bufferRow1.WriteString("│ ")
		bufferRow2.WriteString("│ ")
		bufferRow3.WriteString("│ ")
		bufferRow4.WriteString("│ ")

		// expected format like "2019-10-03Z""
		t, err := time.Parse(layoutMetofficeDate, day.Value)
		if err != nil {
			return ""
		}

		strDate := strconv.Itoa(t.Day())
		if len(strDate) == 1 {
			strDate = " " + strDate
		}

		// Row 1, Column 1: date
		bufferRow1.WriteString(strDate + " ")

		// Row 2, Column 1: Week day (3 first letters)
		bufferRow2.WriteString(t.Month().String()[0:3])

		// Row 3, Column 1: Week day (3 first letters)
		bufferRow3.WriteString(t.Weekday().String()[0:3])

		bufferRow1.WriteString(" │ ")
		bufferRow2.WriteString(" │ ")
		bufferRow3.WriteString(" │ ")

		// Weather icon
		weatherType := 5
		if wt, err := strconv.Atoi(day.Rep[0]["W"]); err == nil {
			weatherType = wt
		}
		bufferRow4.WriteRune(mapWeatherTypes[weatherType].icon)
		bufferRow4.WriteString("  │")
		//bufferRow4.WriteString(mapWeatherTypes[weatherType].name)
		compensateSpaces(&bufferRow4)

		// Row 1, column 2: max day temperature
		bufferRow1.WriteString("T: ")
		bufferRow1.WriteString(day.Rep[0]["Dm"])
		bufferRow1.WriteString("˚C (")
		bufferRow1.WriteString(day.Rep[1]["Nm"])
		bufferRow1.WriteString("˚C)")
		compensateSpaces(&bufferRow1)

		// Row 2, Column 2: max wind speed
		bufferRow2.WriteString("W: ")
		bufferRow2.WriteString(day.Rep[0]["Gn"])
		bufferRow2.WriteString("m/h (")
		bufferRow2.WriteString(day.Rep[1]["Gm"])
		bufferRow2.WriteString("m/h)")
		compensateSpaces(&bufferRow2)

		// Row 3, column 2: rain probability
		bufferRow3.WriteString("R: ")
		bufferRow3.WriteString(day.Rep[0]["PPd"])
		bufferRow3.WriteString("% (")
		bufferRow3.WriteString(day.Rep[1]["PPn"])
		bufferRow3.WriteString("%)")
		compensateSpaces(&bufferRow3)

		bufferRow1.WriteString(" │")
		bufferRow2.WriteString(" │")
		bufferRow3.WriteString(" │")
		bufferRow4.WriteString("│")

		buffer.Write(bufferRow4.Bytes())
		buffer.WriteRune('\n')
		buffer.Write(bufferRow1.Bytes())
		buffer.WriteRune('\n')
		buffer.Write(bufferRow2.Bytes())
		buffer.WriteRune('\n')
		buffer.Write(bufferRow3.Bytes())
		buffer.WriteRune('\n')
		if i < 4 {
			buffer.WriteString("├─────┼────────────────────┤ \n")
		}

		bufferRow1.Reset()
		bufferRow2.Reset()
		bufferRow3.Reset()
		bufferRow4.Reset()
	}

	buffer.WriteString("╰─────┴────────────────────╯ \n```\n")

	return buffer.String()
}

func compensateSpaces(bfr *bytes.Buffer) {
	maxLen := len([]rune(vertTopLine))
	for {
		if len([]rune(bfr.String())) >= maxLen {
			break
		}
		(*bfr).WriteRune(' ')
	}
}

func printDetailedPlotsForADay(data []map[string]string, keyFromMap, unit string, isRound bool) string {
	var buffer bytes.Buffer

	buffer.WriteString("```\n")
	buffer.WriteString("  " + unit)
	buffer.WriteString("\n")

	// kinda, dirty hack. If the width is set to custom value, asciigraph tries
	// to normalize values so they should be rendered accordingly within ascii symbols.
	// In this case values could have decimal values, distorting our graph. However if we
	// do not set a custom width = 3 then a plot will be too short. Within this hack we
	// simply add two more "pixels" in between each values, "stretching" graph.
	multiplier := 3

	temp3Hourly := make([]float64, len(data)*multiplier)

	// remember if a vertical axe has a long value (that is more than 2 symbols),
	// this affects to the bottom line compensation
	hasLongValue := false

	for i, mapHour := range data {
		if intT, err := strconv.Atoi(mapHour[keyFromMap]); err != nil {
			temp3Hourly[i*multiplier] = 0.0
			temp3Hourly[(i*multiplier)+1] = 0.0
			temp3Hourly[(i*multiplier)+2] = 0.0
		} else {
			var fT float64
			if isRound {
				fT = roundToTens(intT)
			} else {
				fT = float64(intT)
			}
			if fT >= 100 {
				hasLongValue = true
			}
			temp3Hourly[i*multiplier] = fT
			temp3Hourly[(i*multiplier)+1] = fT
			temp3Hourly[(i*multiplier)+2] = fT
		}
	}

	// -1 because if all the values are the same, there is panic because of division by 0 inside asciigraph
	lastIndex := len(temp3Hourly) - 1
	temp3Hourly[lastIndex] = temp3Hourly[lastIndex] - 1

	graph := asciigraph.Plot(temp3Hourly)

	// one more dirty hack, I know...
	replacement := ""
	if isRound {
		// oh.... So, even if we round to ten's, asciigraph renders vertical scale for each percent,
		// so 100 percent plot looks huge. And no way to safely shrimp it. I found a hack, that
		// we round and divide by 10, having max plot with 10 lines, and then simply draw artificial "0"
		// simulating multiplying by 10. Sorry :(
		replacement = "0"
	}
	graph = strings.Replace(graph, ".00", replacement, -1)

	buffer.WriteString(graph)
	if len(data) == 8 {
		compensation := ""
		if hasLongValue {
			compensation = " "
		}
		// draw the bottom line only if the day forecast is full:
		// for the current day the first temperature may be started not with 12:00am, but
		// with current day and the bottom line should show only the rest of hours for current day/
		// TODO: improve that and show proper hours left for today
		buffer.WriteString("\n" + compensation + "    └┬──┬──┬──┬──┬──┬──┬──┬")
		buffer.WriteString("\n" + compensation + "     0am   6am   12am  6pm")
	}
	buffer.WriteString("\n```\n")

	return buffer.String()
}

func roundToTens(raw int) float64 {
	return math.Round(float64(raw) / 10)
}
