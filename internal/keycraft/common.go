package keycraft

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

// IfThen returns `a` if the condition is true, otherwise returns `b`.
// Both `a` and `b` are always evaluated before the function is called,
// so avoid using it with expensive operations or values that may be invalid.
func IfThen[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

// WithDefault returns the value for the given key in the map `m` if it exists,
// otherwise returns the provided default value `defVal`.
// Useful for safe map access with a fallback.
func WithDefault[K comparable, V any](m map[K]V, key K, defVal V) V {
	if val, exists := m[key]; exists {
		return val
	}
	return defVal
}

// Must unwraps the value `val` if `err` is nil.
// If `err` is non-nil, it panics. This is useful for simplifying code where
// errors are unexpected or should be fatal (e.g., parsing constants or test setup).
func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

// Must0 panics if the provided error is non-nil.
// This is useful for simplifying code where only an error is returned
// and failures should be considered fatal.
func Must0(err error) {
	if err != nil {
		panic(err)
	}
}

// Pair represents a generic key/value pair used throughout the codebase
// for temporary key-index or key-value collections.
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

// CountPair represents a key/count pair extracted from a map[K]uint64.
type CountPair[K comparable] struct {
	Key   K
	Count uint64
}

// SortedMap returns a slice of key-value pairs from a map, sorted in descending order by count.
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

// CloseFile closes a file and logs any error that occurs.
func CloseFile(file *os.File) {
	if err := file.Close(); err != nil {
		log.Printf("Error closing file: %v", err)
	}
}

// MustFprint writes arguments to the given writer, logging and exiting on error.
// It simplifies error handling for fmt.Fprint calls where failures are critical
// and should halt execution.
func MustFprint(w io.Writer, args ...interface{}) {
	if _, err := fmt.Fprint(w, args...); err != nil {
		log.Fatalf("Fprint failed: %v", err)
	}
}

// mustFprintln writes a newline-terminated string of arguments to the given writer,
// logging and exiting on error. It simplifies error handling for fmt.Fprintln calls
// where failures are critical and should halt execution.
func MustFprintln(w io.Writer, args ...interface{}) {
	if _, err := fmt.Fprintln(w, args...); err != nil {
		log.Fatalf("Fprintln failed: %v", err)
	}
}

// MustFprintf writes a formatted string to the given writer, logging and exiting
// on error. It simplifies error handling for fmt.Fprintf calls where failures
// are critical and should halt execution.
func MustFprintf(w io.Writer, format string, args ...interface{}) {
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		log.Fatalf("Fprintf failed: %v", err)
	}
}

// FlushWriter flushes the buffered writer and logs any error that occurs.
func FlushWriter(writer *bufio.Writer) {
	if err := writer.Flush(); err != nil {
		log.Printf("Error flushing writer: %v", err)
	}
}
