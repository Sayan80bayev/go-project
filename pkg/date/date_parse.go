package date

import "time"

func ParseDate(dateStr string) (time.Time, error) {
	layout := "02.01.2006" // Go's reference layout for DD.MM.YYYY
	return time.Parse(layout, dateStr)
}
