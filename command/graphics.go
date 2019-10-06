package command

import (
	"bytes"
	"github.com/guptarohit/asciigraph"
	"github.com/w32blaster/bot-weather-watcher/structs"
	"strconv"
	"strings"
	"time"
)

const (
	maxSymbolsInRow     = 25 // without the last vertical bar
	vertTopLine         = "╭─────┬───────────────────"
	layoutMetofficeDate = "2006-01-02Z"
)

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

	for i, day := range days {

		bufferRow1.WriteString("│ ")
		bufferRow2.WriteString("│ ")
		bufferRow3.WriteString("│ ")

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

		bufferRow3.WriteString("R: ")
		bufferRow3.WriteString(day.Rep[0]["PPd"])
		bufferRow3.WriteString("% (")
		bufferRow3.WriteString(day.Rep[1]["PPn"])
		bufferRow3.WriteString("%)")
		compensateSpaces(&bufferRow3)

		bufferRow1.WriteString(" │")
		bufferRow2.WriteString(" │")
		bufferRow3.WriteString(" │")

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

func printDetailedPlotsForADay(data []map[string]string, keyFromMap, unit string) string {
	var buffer bytes.Buffer

	buffer.WriteString("```")
	buffer.WriteString(unit)
	buffer.WriteString("\n ")

	// kinda, dirty hack. If the width is set to custom value, asciigraph tries
	// to normalize values so they should be rendered accordingly within ascii symbols.
	// In this case values could have decimal values, distorting our graph. However if we
	// do not set a custom width = 3 then a plot will be too short. Within this hack we
	// simply add two more "pixels" in between each values, "stretching" graph.
	multiplier := 3

	temp3Hourly := make([]float64, len(data)*multiplier)
	for i, mapHour := range data {
		if intT, err := strconv.Atoi(mapHour[keyFromMap]); err != nil {
			temp3Hourly[i*multiplier] = 0.0
			temp3Hourly[(i*multiplier)+1] = 0.0
			temp3Hourly[(i*multiplier)+2] = 0.0
		} else {
			fT := float64(intT)
			temp3Hourly[i*multiplier] = fT
			temp3Hourly[(i*multiplier)+1] = fT
			temp3Hourly[(i*multiplier)+2] = fT
		}
	}
	graph := asciigraph.Plot(temp3Hourly)
	// one more dirty hack, I know...
	graph = strings.Replace(graph, ".00", "", -1)

	buffer.WriteString(graph)
	if len(data) == 8 {
		// draw the bottom line only if the day forecast is full:
		// for the current day the first temperature may be started not with 12:00am, but
		// with current day and the bottom line should show only the rest of hours for current day/
		// TODO: improve that and show proper hours left for today
		buffer.WriteString("\n    └┬──┬──┬──┬──┬──┬──┬──┬")
		buffer.WriteString("\n     0am   6am   12am  6pm")
	}
	buffer.WriteString("\n```\n")

	return buffer.String()
}
