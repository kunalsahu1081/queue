package utility

func NormalizeInt(val int) int {
	if val == nil {
		return 0
	}

	return val
}

func NormalizeTime(val time.Time) time.Time {

	if val == nil {
		return time.Now()
	}

	return val

}