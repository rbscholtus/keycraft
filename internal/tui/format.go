package tui

import "fmt"

// Comma formats an integer with comma separators.
func Comma[T ~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64](v T) string {
	// Convert to uint64 for processing
	val := uint64(v)

	// Calculate the number of digits and commas needed.
	var count byte
	for n := val; n != 0; n = n / 10 {
		count++
	}
	count += (count - 1) / 3

	// Create an output slice to hold the formatted number.
	output := make([]byte, count)
	j := len(output) - 1

	// Populate the output slice with digits and commas.
	var counter byte
	for val > 9 {
		output[j] = byte(val%10) + '0'
		val = val / 10
		j--
		if counter == 2 {
			counter = 0
			output[j] = ','
			j--
		} else {
			counter++
		}
	}

	output[j] = byte(val) + '0'

	return string(output)
}

// Thousands formats a uint64 count using comma separators, or falls back to %v for other types.
func Thousands(val any) string {
	if number, ok := val.(uint64); ok {
		return Comma(number)
	}
	return fmt.Sprintf("%v", val)
}

// Fraction formats a float64 with two decimals, or falls back to %v for other types.
func Fraction(val any) string {
	if number, ok := val.(float64); ok {
		return fmt.Sprintf("%.2f", number)
	}
	return fmt.Sprintf("%v", val)
}

// Angle formats numeric values as degree strings.
func Angle(val any) string {
	switch v := val.(type) {
	case float32, float64:
		return fmt.Sprintf("%.1f°", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d°", v)
	default:
		return fmt.Sprint(val)
	}
}

// Percentage formats a fractional value (0..1) as a percentage with two decimals.
func Percentage(val any) string {
	if number, ok := val.(float64); ok {
		return fmt.Sprintf("%.2f%%", 100*number)
	}
	return fmt.Sprintf("%v", val)
}

// Percentage3 formats a fractional value (0..1) as a percentage with three decimals.
func Percentage3(val any) string {
	if number, ok := val.(float64); ok {
		return fmt.Sprintf("%.3f%%", 100*number)
	}
	return fmt.Sprintf("%v", val)
}
