package keycraft

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Unigram represents a single character (1-gram) in a text corpus.
type Unigram rune

// String returns the string representation of the unigram.
func (u Unigram) String() string {
	return string(u)
}

// MarshalText implements the encoding.TextMarshaler interface for Unigram.
func (u Unigram) MarshalText() ([]byte, error) {
	return []byte(string(u)), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for Unigram.
func (u *Unigram) UnmarshalText(text []byte) error {
	runes := []rune(string(text))
	if len(runes) != 1 {
		return fmt.Errorf("invalid Unigram length: %d", len(runes))
	}
	*u = Unigram(runes[0])
	return nil
}

// Bigram represents a sequence of two characters (2-gram) in a text corpus.
type Bigram [2]rune

// String returns the string representation of the bigram.
func (b Bigram) String() string {
	return string(b[:])
}

// MarshalText implements the encoding.TextMarshaler interface for Bigram.
func (b Bigram) MarshalText() ([]byte, error) {
	return []byte(string(b[:])), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for Bigram.
func (b *Bigram) UnmarshalText(text []byte) error {
	runes := []rune(string(text))
	if len(runes) != 2 {
		return fmt.Errorf("invalid Bigram length: %d", len(runes))
	}
	b[0], b[1] = runes[0], runes[1]
	return nil
}

// Trigram represents a sequence of three characters (3-gram) in a text corpus.
type Trigram [3]rune

// String returns the string representation of the trigram.
func (t Trigram) String() string {
	return string([]rune{t[0], t[1], t[2]})
}

// MarshalText implements the encoding.TextMarshaler interface for Trigram.
func (t Trigram) MarshalText() ([]byte, error) {
	return []byte(string([]rune{t[0], t[1], t[2]})), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for Trigram.
func (t *Trigram) UnmarshalText(text []byte) error {
	runes := []rune(string(text))
	if len(runes) != 3 {
		return fmt.Errorf("invalid Trigram length: %d", len(runes))
	}
	t[0], t[1], t[2] = runes[0], runes[1], runes[2]
	return nil
}

// Skipgram represents a skipgram, consisting of the first and last characters of a 3-character sequence.
type Skipgram [2]rune

// String returns the string representation of the skipgram.
func (b Skipgram) String() string {
	return string(b[:])
}

// MarshalText implements the encoding.TextMarshaler interface for Skipgram.
func (s Skipgram) MarshalText() ([]byte, error) {
	return []byte(string(s[:])), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for Skipgram.
func (s *Skipgram) UnmarshalText(text []byte) error {
	runes := []rune(string(text))
	if len(runes) != 2 {
		return fmt.Errorf("invalid Skipgram length: %d", len(runes))
	}
	s[0], s[1] = runes[0], runes[1]
	return nil
}

// Corpus represents a text corpus used for analysing character-level statistics.
// It tracks frequency counts and totals for different types of n-grams:
//
//   - Unigrams: single characters
//   - Bigrams: consecutive two-character sequences
//   - Trigrams: consecutive three-character sequences
//   - Skipgrams: two-character sequences formed from the first and last character of a three-character window
//
// Each map stores the frequency of the corresponding n-gram, while the associated
// Total*Count fields store the aggregate counts across the entire corpus.
// The Name field identifies the corpus and is used in filenames when saving to or loading from JSON.
type Corpus struct {
	// Name identifies the corpus and is also used when generating filenames for JSON serialization.
	Name string

	// Unigrams maps each individual character (unigram) to the number of times it occurs in the corpus.
	Unigrams map[Unigram]uint64
	// TotalUnigramsCount is the total number of unigram instances observed across the entire corpus.
	TotalUnigramsCount uint64

	// Bigrams maps each pair of consecutive characters (bigram) to the number of times it occurs in the corpus.
	Bigrams map[Bigram]uint64
	// TotalBigramsCount is the total number of bigram instances observed across the entire corpus.
	TotalBigramsCount uint64

	// Trigrams maps each sequence of three consecutive characters (trigram) to the number of times it occurs in the corpus.
	Trigrams map[Trigram]uint64
	// TotalTrigramsCount is the total number of trigram instances observed across the entire corpus.
	TotalTrigramsCount uint64

	// Skipgrams maps each skipgram (two characters formed by skipping the middle character in a three-character window)
	// to the number of times it occurs in the corpus.
	Skipgrams map[Skipgram]uint64
	// TotalSkipgramsCount is the total number of skipgram instances observed across the entire corpus.
	TotalSkipgramsCount uint64
}

// NewCorpus creates and returns a new empty Corpus with the given name.
func NewCorpus(name string) *Corpus {
	return &Corpus{
		Name:      name,
		Unigrams:  make(map[Unigram]uint64),
		Bigrams:   make(map[Bigram]uint64),
		Trigrams:  make(map[Trigram]uint64),
		Skipgrams: make(map[Skipgram]uint64),
	}
}

// StringSorted returns a string representation of the corpus, listing n-grams sorted by their counts.
// The limit parameter controls how many top n-grams to include for each n-gram type.
// If limit is zero or negative, all n-grams are included.
func (c *Corpus) StringSorted(limit int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Corpus: %s\n", c.Name))

	if limit <= 0 {
		limit = math.MaxInt32
	}

	// Print unigrams
	sb.WriteString("Unigrams:\n")
	uc := SortedMap(c.Unigrams)
	for i := 0; i < limit && i < len(uc); i++ {
		sb.WriteString(fmt.Sprintf("%s: %d\n", uc[i].Key, uc[i].Count))
	}

	// Print bigrams
	sb.WriteString("Bigrams:\n")
	bc := SortedMap(c.Bigrams)
	for i := 0; i < limit && i < len(bc); i++ {
		sb.WriteString(fmt.Sprintf("%s: %d\n", bc[i].Key, bc[i].Count))
	}

	// Print trigrams
	sb.WriteString("Trigrams:\n")
	tc := SortedMap(c.Trigrams)
	for i := 0; i < limit && i < len(tc); i++ {
		sb.WriteString(fmt.Sprintf("%s: %d\n", tc[i].Key, tc[i].Count))
	}

	// Print skipgrams
	sb.WriteString("Skipgrams:\n")
	sc := SortedMap(c.Skipgrams)
	for i := 0; i < limit && i < len(sc); i++ {
		sb.WriteString(fmt.Sprintf("%s: %d\n", sc[i].Key, sc[i].Count))
	}

	return sb.String()
}

// String returns a string representation of the corpus showing the top 10 bigrams by default.
func (c *Corpus) String() string {
	return c.StringSorted(10)
}

// NewCorpusFromFile creates a new Corpus with the given name by loading data from the specified text file.
// It attempts to load from a cached JSON file if it exists and is newer than the source text file.
// If no valid cache is found, it loads from the text file and saves a JSON cache for future use.
func NewCorpusFromFile(name, path string) (*Corpus, error) {
	// Compute the JSON filename in the same directory as filename
	jsonPath := filepath.Join(path + ".json")

	// Check if JSON file exists and is newer than source file, or if source file is missing
	jsonInfo, jsonErr := os.Stat(jsonPath)
	srcInfo, srcErr := os.Stat(path)
	if jsonErr == nil && (os.IsNotExist(srcErr) || (srcErr == nil && jsonInfo.ModTime().After(srcInfo.ModTime()))) {
		return LoadJSON(jsonPath)
	}

	// Otherwise, load from the text file and save JSON cache
	c := NewCorpus(name)
	if err := c.loadFromFile(path); err != nil {
		return nil, err
	}
	if err := c.SaveJSON(jsonPath); err != nil {
		return nil, err
	}

	return c, nil
}

// addUnigram increments the count of the given unigram in the corpus.
func (c *Corpus) addUnigram(r rune) {
	u := Unigram(r)
	c.Unigrams[u]++
	c.TotalUnigramsCount++
}

// addBigram increments the count of the given bigram in the corpus.
func (c *Corpus) addBigram(r1, r2 rune) {
	bigram := Bigram{r1, r2}
	c.Bigrams[bigram]++
	c.TotalBigramsCount++
}

// addTrigram increments the count of the given trigram in the corpus.
func (c *Corpus) addTrigram(r1, r2, r3 rune) {
	trigram := Trigram{r1, r2, r3}
	c.Trigrams[trigram]++
	c.TotalTrigramsCount++
}

// addSkipgram increments the count of the given skipgram in the corpus.
func (c *Corpus) addSkipgram(r1, r2 rune) {
	skipgram := Skipgram{r1, r2}
	c.Skipgrams[skipgram]++
	c.TotalSkipgramsCount++
}

// addText processes the input text string, converting it to lowercase and adding all unigrams,
// bigrams, trigrams, and skipgrams to the corpus. N-grams that contain whitespace are skipped.
func (c *Corpus) addText(text string) {
	text = strings.ToLower(text)
	var prev1, prev2 rune
	for _, r := range text {
		if unicode.IsSpace(r) {
			prev1 = 0
			prev2 = 0
			continue
		}

		c.addUnigram(r)

		// Add bigram if previous rune exists
		if prev1 != 0 {
			c.addBigram(prev1, r)

			// Add trigram and skipgram if two previous runes exist
			if prev2 != 0 {
				c.addTrigram(prev2, prev1, r)
				c.addSkipgram(prev2, r)
			}
		}

		prev2 = prev1
		prev1 = r
	}
}

// loadFromFile reads the given text file line by line and adds the text content to the corpus.
// Empty or whitespace-only lines are skipped.
func (c *Corpus) loadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer CloseFile(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		c.addText(line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// LoadJSON loads a Corpus from the specified JSON file.
func LoadJSON(jsonPath string) (*Corpus, error) {
	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		return nil, err
	}
	defer CloseFile(jsonFile)

	var c Corpus
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

// SaveJSON saves the corpus data as a JSON file in the specified directory using the corpus name as filename.
func (c *Corpus) SaveJSON(jsonPath string) error {
	file, err := os.Create(jsonPath)
	if err != nil {
		return err
	}
	defer CloseFile(file)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return err
	}
	return nil
}
