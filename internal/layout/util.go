// Package comma provides functions for formatting numbers with commas.
package layout

// Comma returns a string representation of the given number with commas.
func Comma(v uint64) string {
	// Calculate the number of digits and commas needed.
	var count byte = 0
	for n := v; n != 0; n = n / 10 {
		count++
	}
	count += (count - 1) / 3

	// Create an output slice to hold the formatted number.
	output := make([]byte, count)
	j := len(output) - 1

	// Populate the output slice with digits and commas.
	var counter byte = 0
	for v > 9 {
		output[j] = byte(v%10) + '0'
		v = v / 10
		j--
		if counter == 2 {
			counter = 0
			output[j] = ','
			j--
		} else {
			counter++
		}
	}

	output[j] = byte(v) + '0'

	return string(output)
}

// Pair is a generic key-value pair struct.
type Pair[K comparable, V any] struct {
	// Key is the key of the pair.
	Key K
	// Value is the value associated with the key.
	Value V
}

// ifThen returns a value based on the given condition.
func ifThen[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}
