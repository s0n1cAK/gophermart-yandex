package luhn

import "unicode"

func CalculateLuhn(number string) int {
	sum := 0
	double := len(number)%2 == 0

	for _, r := range number {
		if !unicode.IsDigit(r) {
			return -1
		}

		n := int(r - '0')
		if double {
			n = n * 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		double = !double
	}

	return (10 - (sum % 10)) % 10
}

func Valid(number string) bool {
	if len(number) == 0 {
		return false
	}

	sum := 0
	double := len(number)%2 == 0

	for _, r := range number {
		if !unicode.IsDigit(r) {
			return false
		}

		n := int(r - '0')
		if double {
			n = n * 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		double = !double
	}

	return sum%10 == 0
}
