package bcdate

import "time"

type BCDate uint

func CurrentMonth() BCDate {
	n := time.Now()

	return BCDate(int(n.Year())*10000 + int(n.Month())*100)
}

func Epoch() BCDate {
	return 0
}

func Oldest(a BCDate, b BCDate) BCDate {
	if a < b {
		return a
	} else {
		return b
	}
}
func Latest(a BCDate, b BCDate) BCDate {
	if a > b {
		return a
	} else {
		return b
	}
}
