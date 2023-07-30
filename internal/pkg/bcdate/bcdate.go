package bcdate

import (
	"fmt"
	"time"
)

type BCDate uint

func CurrentMonth() BCDate {
	n := time.Now()

	return BCDate(int(n.Year())*10000 + int(n.Month())*100)
}

func Epoch() BCDate {
	return 0
}

func Never() BCDate {
	// If this survives until the year 40k, we should worry about galaxy-wide war, not budgeting
	return 400000000
}

func Oldest(a BCDate, b BCDate) BCDate {
	if a < b {
		return a
	} else {
		return b
	}
}
func Latest(a BCDate, b BCDate) BCDate {
	if (a/100)%100 == 1 {
		a = (a/10000)*10000 - 10000 + 100 + a%100
	}
	return a
}

func (a BCDate) PrevMonth() BCDate {
	return a - 100
}
func (a BCDate) NextMonth() BCDate {
	a += 100
	if (a/100)%100 > 12 {
		a = (a/10000)*10000 + 10100 + a%100
	}
	return a
}

func (a BCDate) FmtDate() string {
	day := a % 100
	mon := ((a - day) % 10000) / 100
	yer := (a - day - mon) / 10000

	return fmt.Sprintf("%04d-%02d-%02d", yer, mon, day)
}

func (a BCDate) FmtMonth() string {
	day := a % 100
	mon := ((a - day) % 10000) / 100
	yer := (a - day - mon) / 10000

	return fmt.Sprintf("%04d-%02d", yer, mon)
}
