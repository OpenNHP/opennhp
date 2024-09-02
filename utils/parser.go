package utils

import (
	"fmt"
	"strconv"
)

func ParseBool(v int) bool {
	return v == 1
}

func ParseInt(v string) int {
	r, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return r
}

func ParseUInt64(v string) uint64 {
	r, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0
	}
	return r
}

func ParseInt64ToInt(v int64) int {
	res, err := strconv.Atoi(fmt.Sprintf("%d", v))
	if err != nil {
		return 0
	}
	return res
}
