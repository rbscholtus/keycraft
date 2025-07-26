// Package layout provides common structs and utility functions.
package layout

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
)

// Unigram represents a 1-character sequence
type Unigram rune

// String returns a string representation of the unigram
func (u Unigram) String() string {
	return string(u)
}

// UnigramCount represents a unigram and its count
type UnigramCount struct {
	Unigram Unigram
	Count   uint64
}

// Bigram represents a 2-character sequence
type Bigram [2]rune

// String returns a string representation of the bigram
func (b Bigram) String() string {
	return string([]rune{b[0], b[1]})
}

// BigramCount represents a bigram and its count
type BigramCount struct {
	Bigram Bigram
	Count  uint64
}

// Trigram represents a 3-character sequence
type Trigram [3]rune

// String returns a string representation of the trigram
func (t Trigram) String() string {
	return string([]rune{t[0], t[1], t[2]})
}

// TrigramCount represents a trigram and its count
type TrigramCount struct {
	Trigram Trigram
	Count   uint64
}

// Comma returns a string representation of the given number with commas.
func Comma(v uint64) string {
	// Calculate the number of digits and commas needed.
	var count byte
	for n := v; n != 0; n = n / 10 {
		count++
	}
	count += (count - 1) / 3

	// Create an output slice to hold the formatted number.
	output := make([]byte, count)
	j := len(output) - 1

	// Populate the output slice with digits and commas.
	var counter byte
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

func Fraction(val any) string {
	if number, ok := val.(float32); ok {
		return fmt.Sprintf("%.2f", number)
	}
	return fmt.Sprintf("%v", val)
}

func Percentage(val any) string {
	if number, ok := val.(float32); ok {
		return fmt.Sprintf("%.2f%%", 100*number)
	}
	return fmt.Sprintf("%v", val)
}

func Thousands(val any) string {
	if number, ok := val.(uint64); ok {
		return Comma(number)
	}
	return fmt.Sprintf("%v", val)
}

// IfThen returns a value based on the given condition.
func IfThen[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

// Pair is a generic key-value pair struct.
type Pair[K comparable, V any] struct {
	// Key is the key of the pair.
	Key K
	// Value is the value associated with the key.
	Value V
}

// CountPair represents a key-value pair from a map, where the value is a count
type CountPair[K comparable] struct {
	Key   K      // the key from the map
	Count uint64 // the count associated with the key
}

// SortedMap returns a slice of key-value pairs from a map, sorted in descending order by count
func SortedMap[K comparable](m map[K]uint64) []CountPair[K] {
	if m == nil {
		return []CountPair[K]{}
	}

	pairs := make([]CountPair[K], 0, len(m))
	for k, v := range m {
		pairs = append(pairs, CountPair[K]{k, v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	return pairs
}

// Closes a file and prints the error, if any
func CloseFile(file *os.File) {
	if err := file.Close(); err != nil {
		log.Printf("Error closing file: %v", err)
	}
}

// Flushes the writer and prints the error, if any
func FlushWriter(writer *bufio.Writer) {
	if err := writer.Flush(); err != nil {
		log.Printf("Error flushing writer: %v", err)
	}
}

func Abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
