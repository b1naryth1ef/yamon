package util

import (
	"strconv"
)

func FilterRepeatingSpaces(parts []string) []string {
	result := []string{}
	for _, i := range parts {
		if i == "" {
			continue
		}
		result = append(result, i)
	}
	return result
}

func ParseNumber(i string) uint64 {
	v, err := strconv.ParseInt(i, 10, 64)
	if err != nil {
		return 0
	}
	return uint64(v)
}

func ParseFloat(i string) float64 {
	v, err := strconv.ParseFloat(i, 64)
	if err != nil {
		return 0
	}
	return v
}
