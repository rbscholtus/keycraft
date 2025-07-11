// internal/corpus/corpus.go
package corpus

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

// Bigram represents a 2-character sequence
type Bigram [2]rune

// String returns a string representation of the bigram
func (b Bigram) String() string {
	return string([]rune{b[0], b[1]})
}

// Corpus represents a corpus of text
type Corpus struct {
	Name    string
	Bigrams map[Bigram]int
}

// NewCorpus creates a new corpus
func NewCorpus(name string) *Corpus {
	return &Corpus{
		Name:    name,
		Bigrams: make(map[Bigram]int),
	}
}

// StringSimple returns a simple string representation of the corpus
func (c *Corpus) StringSimple() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Corpus: %s\n", c.Name))
	for bigram, count := range c.Bigrams {
		sb.WriteString(fmt.Sprintf("%s: %d\n", bigram.String(), count))
	}
	return sb.String()
}

// StringSorted returns a string representation of the corpus sorted by bigram count
func (c *Corpus) StringSorted(limit int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Corpus: %s\n", c.Name))

	type bigramCount struct {
		bigram Bigram
		count  int
	}
	bc := make([]bigramCount, 0, len(c.Bigrams))
	for bigram, count := range c.Bigrams {
		bc = append(bc, bigramCount{bigram, count})
	}

	sort.Slice(bc, func(i, j int) bool {
		return bc[i].count > bc[j].count
	})

	if limit > len(bc) {
		limit = len(bc)
	}

	for i := 0; i < limit; i++ {
		sb.WriteString(fmt.Sprintf("%s: %d\n", bc[i].bigram.String(), bc[i].count))
	}

	return sb.String()
}

// String returns a string representation of the top 30 bigrams in the corpus
func (c *Corpus) String() string {
	return c.StringSorted(30)
}

// NewFromFile creates a new corpus loaded from a file
func NewFromFile(name string, filename string) (*Corpus, error) {
	c := NewCorpus(name)
	err := c.loadFromFile(filename)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// AddBigram adds a bigram to the corpus
func (c *Corpus) AddBigram(bigram Bigram) {
	c.Bigrams[bigram]++
}

// AddText adds Bigrams in the text to the corpus, skipping bigrams with a space
func (c *Corpus) AddText(text string) {
	text = strings.ToLower(text)
	runes := []rune(text)
	for i := 0; i < len(runes)-1; i++ {
		if !unicode.IsSpace(runes[i]) && !unicode.IsSpace(runes[i+1]) {
			bigram := Bigram{runes[i], runes[i+1]}
			c.AddBigram(bigram)
		}
	}
}

// loadFromFile loads text from a file into the corpus
func (c *Corpus) loadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

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
