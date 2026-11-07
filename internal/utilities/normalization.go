package utility

import "time"

func NormalizeInt(val int) int {
	if val < 0 {
		return 0
	}

	return val
}

func NormalizeTime(val time.Time) time.Time {

	if val.IsZero() {
		return time.Now()
	}

	return val

}
