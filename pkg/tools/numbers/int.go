package numbers

import "strconv"

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
