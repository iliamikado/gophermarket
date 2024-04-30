package router

import (
	"strconv"
	"strings"
)

func checkOrderNumber(orderNumber string) bool {
	digits := strings.Split(orderNumber, "")
	var parity = len(digits) % 2
	var sum = 0
	for i := 0; i < len(digits); i++ {
		d, err := strconv.Atoi(digits[i])
		if err != nil {
			return false
		}
		if i % 2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum % 10 == 0
}