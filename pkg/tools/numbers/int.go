package numbers

import (
	"math"
	"strconv"
)

// StrToIntMust return 0 if sting empty
func StrToIntMust(str string) int {
	if str == "" {
		return 0
	}
	r, err := strconv.Atoi(str)

	if err != nil {
		panic(err)
	}

	return r
}

func NumberRoundTo(number, to int) int {
	if number == 0 {
		return to
	}
	if math.Mod(float64(number), float64(to)) != 0 {
		if to > 0 {
			number = number - int(math.Mod(float64(number), float64(to))) + to
		} else {
			number = number - int(math.Mod(float64(number), float64(to)))
		}
	}
	return number
}
