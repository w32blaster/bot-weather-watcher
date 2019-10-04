package command

import (
	"bytes"
	"github.com/w32blaster/bot-weather-watcher/structs"
	"strconv"
	"time"
)

const (
	maxSymbolsInRow = 25 // without the last vertical bar
	vertTopLine     = "╭─────┬───────────────────"
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
		layout := "2006-01-02Z"
		t, err := time.Parse(layout, day.Value)
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
