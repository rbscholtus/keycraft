// Package layout provides structs and algorithms for representing a text corpus.
package layout

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
	"unicode"
)

// Corpus represents a corpus of text
type Corpus struct {
	Name                 string
	Unigrams             map[Unigram]uint64 // map of unigrams to their counts
	TotalUnigramsCount   uint64             // total count of unigrams
	TotalUnigramsNoSpace uint64             // total count of unigrams with no spaces
	Bigrams              map[Bigram]uint64  // map of bigrams to their counts
	TotalBigramsCount    uint64             // total count of bigrams
	TotalBigramsNoSpace  uint64             // total count of bigrams with no spaces
	Trigrams             map[Trigram]uint64 // map of trigrams to their counts
	TotalTrigramsCount   uint64             // total count of trigrams
	TotalTrigramsNoSpace uint64             // total count of trigrams with no spaces
}

// NewCorpus creates a new corpus with the given name
func NewCorpus(name string) *Corpus {
	return &Corpus{
		Name:     name,
		Unigrams: make(map[Unigram]uint64),
		Bigrams:  make(map[Bigram]uint64),
		Trigrams: make(map[Trigram]uint64),
	}
}

// StringSorted returns a string representation of the corpus sorted by n-gram count
func (c *Corpus) StringSorted(limit int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Corpus: %s\n", c.Name))

	if limit <= 0 {
		limit = math.MaxInt32
	}

	// Print unigrams
	sb.WriteString("Unigrams:\n")
	uc := SortedMap(c.Unigrams)
	if limit > len(uc) {
		limit = len(uc)
	}

	for i := 0; i < limit && i < len(uc); i++ {
		sb.WriteString(fmt.Sprintf("%s: %d\n", uc[i].Key, uc[i].Count))
	}

	// Print bigrams
	sb.WriteString("Bigrams:\n")
	bc := SortedMap(c.Bigrams)
	if limit > len(bc) {
		limit = len(bc)
	}

	for i := 0; i < limit && i < len(bc); i++ {
		sb.WriteString(fmt.Sprintf("%s: %d\n", bc[i].Key, bc[i].Count))
	}

	// Print trigrams
	sb.WriteString("\nTrigrams:\n")
	tc := SortedMap(c.Trigrams)
	if limit > len(tc) {
		limit = len(tc)
	}

	for i := 0; i < limit && i < len(tc); i++ {
		sb.WriteString(fmt.Sprintf("%s: %d\n", tc[i].Key, tc[i].Count))
	}

	return sb.String()
}

// String returns a string representation of the top 30 bigrams in the corpus
func (c *Corpus) String() string {
	return c.StringSorted(10)
}

// NewCorpusFromFile creates a new corpus loaded from a file
func NewCorpusFromFile(name, filename string) (*Corpus, error) {
	c := NewCorpus(name)
	err := c.loadFromFile(filename)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// AddUnigram adds a unigram to the corpus and increments its count
func (c *Corpus) AddUnigram(u Unigram) {
	c.Unigrams[u]++
	c.TotalUnigramsCount++
	if !unicode.IsSpace(rune(u)) {
		c.TotalUnigramsNoSpace++
	}
}

// AddBigram adds a bigram to the corpus and increments its count
func (c *Corpus) AddBigram(bigram Bigram) {
	c.Bigrams[bigram]++
	c.TotalBigramsCount++
	if !unicode.IsSpace(bigram[0]) && !unicode.IsSpace(bigram[1]) {
		c.TotalBigramsNoSpace++
	}
}

// AddTrigram adds a trigram to the corpus and increments its count
func (c *Corpus) AddTrigram(trigram Trigram) {
	c.Trigrams[trigram]++
	c.TotalTrigramsCount++
	if !unicode.IsSpace(trigram[0]) && !unicode.IsSpace(trigram[1]) && !unicode.IsSpace(trigram[2]) {
		c.TotalTrigramsNoSpace++
	}
}

// AddText adds Unigrams, Bigrams, and Trigrams in the text to the corpus, skipping n-grams with a space
func (c *Corpus) AddText(text string) {
	text = strings.ToLower(text)
	var prev1, prev2 rune
	for _, r := range text {
		// if unicode.IsSpace(r) {
		// 	prev1 = 0
		// 	prev2 = 0
		// 	continue
		// }

		c.AddUnigram(Unigram(r))
		if prev1 != 0 {
			bigram := Bigram{prev1, r}
			c.AddBigram(bigram)
		}
		if prev2 != 0 && prev1 != 0 {
			trigram := Trigram{prev2, prev1, r}
			c.AddTrigram(trigram)
		}
		prev2 = prev1
		prev1 = r
	}
}

// loadFromFile loads text from a file into the corpus
func (c *Corpus) loadFromFile(filename string) error {
	file, err := os.Open(filename)
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
		c.AddText(line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
