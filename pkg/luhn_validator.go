package pkg

// ValidateString проверяет переданную строку на корректность, используя
// алгоритм Луна. Данная функция ожидает, что переданная строка не содержит
// разделителей (пробел, тире и т.д.)
//
// Если переданная строка не содержит ошибок, возвращает true. Иначе - false.
func ValidateString(input string) bool {
	// специальный случай: если вводная строка пуста - проверять нечего
	if len(input) == 0 {
		return false
	}

	sum := 0
	nextDouble := false

	for i := len(input) - 1; i >= 0; i-- {
		digit := int(input[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if nextDouble {
			digit *= 2
		}
		if digit > 9 {
			digit = digit - 9
		}

		sum += digit
		nextDouble = !nextDouble
	}

	return sum%10 == 0
}
