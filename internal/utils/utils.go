package utils

import (
	"log"
	"strconv"
)

func TryConvertToInt(value string) int {
	if result, err := strconv.Atoi(value); err != nil {
		log.Printf("Error converting to integer: %v", err)
		return 0
	} else {
		return result
	}
}
