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
	return string(b[:])
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

// Skipgram represents the first and last character of a 3-character sequence
type Skipgram [2]rune

// String returns a string representation of the skipgram
func (b Skipgram) String() string {
	return string(b[:])
}

// SkipgramCount represents a skipgram and its count
type SkipgramCount struct {
	Skipgram Skipgram
	Count    uint64
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
	if number, ok := val.(float64); ok {
		return fmt.Sprintf("%.2f", number)
	}
	return fmt.Sprintf("%v", val)
}

func Percentage(val any) string {
	if number, ok := val.(float64); ok {
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

func WithDefault[K comparable, V any](m map[K]V, key K, defVal V) V {
	if val, exists := m[key]; exists {
		return val
	}
	return defVal
}

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
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

// Calculate the median of a sorted slice
func Median(sortedData []float64) float64 {
	n := len(sortedData)
	mid := n / 2
	if n%2 == 0 {
		return (sortedData[mid-1] + sortedData[mid]) / 2.0
	} else {
		return sortedData[mid]
	}
}

// Calculate the first and third quartiles
func Quartiles(sortedData []float64) (float64, float64) {
	n := len(sortedData)
	q1 := Median(sortedData[:n/2])
	q3 := Median(sortedData[(n+1)/2:])
	return q1, q3
}

// Robust scaling
func RobustScale(data []float64) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	sortedData := make([]float64, len(data))
	copy(sortedData, data)
	sort.Float64s(sortedData)

	medianValue := Median(sortedData)
	q1, q3 := Quartiles(sortedData)
	iqr := q3 - q1

	if iqr == 0 {
		return make([]float64, len(data))
	}

	scaledData := make([]float64, len(data))
	for i, x := range data {
		scaledData[i] = (x - medianValue) / iqr
	}

	return scaledData
}
