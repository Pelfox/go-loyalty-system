package pkg

import (
	"testing"
)

// TestValidateString_ValidNumbers проверяет корректность работы функции на
// заранее заготовленных числах, которые должны вернуть true.
func TestValidateString_ValidNumbers(t *testing.T) {
	tests := []string{
		"79927398713",
		"4539148803436467",
		"4012888888881881",
		"378282246310005",
	}

	for _, tc := range tests {
		if !ValidateString(tc) {
			t.Errorf("expected %q to be valid", tc)
		}
	}
}

// TestValidateString_InvalidNumbers проверяет корректность работы функции,
// используя заранее неверные числа (все они должны вернуть false).
func TestValidateString_InvalidNumbers(t *testing.T) {
	tests := []string{
		"79927398710",
		"4539148803436466",
		"4012888888881882",
	}

	for _, tc := range tests {
		if ValidateString(tc) {
			t.Errorf("expected %q to be invalid", tc)
		}
	}
}

// TestValidateString_NonDigitInput тестирует функцию на ввод некорректных
// символов в строке (не цифр).
func TestValidateString_NonDigitInput(t *testing.T) {
	tests := []string{
		"",
		"1234 5678",
		"1234-5678",
		"12a34",
		"0000x",
	}

	for _, tc := range tests {
		if ValidateString(tc) {
			t.Errorf("expected %q to be invalid (non-digit input)", tc)
		}
	}
}

// TestValidateString_SingleDigit проверяет специфичные случаи функции для
// строк из одного числа.
func TestValidateString_SingleDigit(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"0", true},
		{"5", false},
		{"9", false},
	}

	for _, tc := range tests {
		if got := ValidateString(tc.input); got != tc.want {
			t.Errorf("ValidateString(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
