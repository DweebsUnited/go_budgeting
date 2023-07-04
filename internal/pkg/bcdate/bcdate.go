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
