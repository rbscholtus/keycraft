package keycraft

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
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

	// Words maps each unique word in the corpus to the number of times it occurs.
	Words map[string]uint64
	// TotalWordsCount is the total number of word instances observed across the entire corpus.
	TotalWordsCount uint64
}

// NewCorpus creates and returns a new empty Corpus with the given name.
func NewCorpus(name string) *Corpus {
	return &Corpus{
		Name:      name,
		Unigrams:  make(map[Unigram]uint64),
		Bigrams:   make(map[Bigram]uint64),
		Trigrams:  make(map[Trigram]uint64),
		Skipgrams: make(map[Skipgram]uint64),
		Words:     make(map[string]uint64),
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

// String returns a string representation of the corpus showing the top 10 n-grams by default.
func (c *Corpus) String() string {
	return c.StringSorted(10)
}

// NewCorpusFromFile creates a new Corpus with the given name by loading data from the specified text file.
// It attempts to load from a cached JSON file if it exists and is newer than the source text file.
// If no valid cache is found, it loads from the text file and saves a JSON cache for future use.
// If forceReload is true, it skips loading from JSON and always rebuilds from text.
func NewCorpusFromFile(name, path string, forceReload bool, coveragePercent float64) (*Corpus, error) {
	// Compute the JSON filename in the same directory as filename
	jsonPath := path + ".json"

	// If forceReload is false, try to load from JSON cache if it exists and is newer than source file
	if !forceReload {
		jsonInfo, jsonErr := os.Stat(jsonPath)
		srcInfo, srcErr := os.Stat(path)
		if jsonErr == nil && (os.IsNotExist(srcErr) || (srcErr == nil && jsonInfo.ModTime().After(srcInfo.ModTime()))) {
			corpus, err := LoadJSON(jsonPath)
			if err != nil {
				return nil, fmt.Errorf("could not load corpus from cache: %w", err)
			}
			return corpus, nil
		}
	}

	// Otherwise, load from the text file and save JSON cache
	c := NewCorpus(name)
	if err := c.loadFromFileWithWords(path, coveragePercent); err != nil {
		return nil, fmt.Errorf("could not load corpus from file: %w", err)
	}
	if err := c.SaveJSON(jsonPath); err != nil {
		return nil, fmt.Errorf("could not save corpus cache: %w", err)
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

// addWord increments the count of the given word in the corpus
//
//nolint:unused
func (c *Corpus) addWord(word string) {
	c.Words[word]++
	c.TotalWordsCount++
}

// addText processes text and extracts n-grams (unigrams, bigrams, trigrams, skipgrams).
// Text is lowercased, and n-grams containing whitespace are skipped (word boundaries reset the window).
//
//nolint:unused
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

// addTextWithWords processes text, extracting both words and n-grams.
// Words are defined as sequences of letters and numbers, with support for apostrophes
// in contractions (e.g., "don't", "she'll") and possessives (e.g., "John's", "users'").
func (c *Corpus) addTextWithWords(text string) {
	text = strings.ToLower(text)

	// Helper function to check if a rune is an apostrophe (ASCII or Unicode)
	isApostrophe := func(r rune) bool {
		return r == '\'' || r == '\u2019' // ' or '
	}

	// Extract words (alphanumeric sequences with apostrophes)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && !isApostrophe(r)
	})

	// Post-process words to trim leading/trailing apostrophes
	for _, word := range words {
		if word == "" {
			continue
		}
		// Trim leading apostrophes (e.g., "'hello" -> "hello")
		word = strings.TrimLeftFunc(word, isApostrophe)
		// Trim trailing apostrophes only if multiple (e.g., "hello''" -> "hello'")
		// Keep single trailing apostrophe for possessives (e.g., "users'")
		for len(word) > 0 && isApostrophe(rune(word[len(word)-1])) {
			if len(word) > 1 && isApostrophe(rune(word[len(word)-2])) {
				word = word[:len(word)-1]
			} else {
				break
			}
		}
		if word != "" {
			c.addWord(word)
		}
	}

	// Extract n-grams using sliding window
	var prev1, prev2 rune
	for _, r := range text {
		if unicode.IsSpace(r) {
			prev1 = 0
			prev2 = 0
			continue
		}

		c.addUnigram(r)

		if prev1 != 0 {
			c.addBigram(prev1, r)

			if prev2 != 0 {
				c.addTrigram(prev2, prev1, r)
				c.addSkipgram(prev2, r)
			}
		}

		prev2 = prev1
		prev1 = r
	}
}

// loadFromFile reads a text file line by line and extracts n-grams.
// Empty or whitespace-only lines are skipped.
//
//nolint:unused // Superseded by loadFromFileWithWords but kept for potential future use.
func (c *Corpus) loadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
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
		return fmt.Errorf("could not read file: %w", err)
	}

	return nil
}

// loadFromFileWithWords loads text from a file, extracting both words and n-grams.
// After loading, prunes the word list to keep only the most frequent words covering
// the specified percentage of total word occurrences.
//
//nolint:unused
func (c *Corpus) loadFromFileWithWords(path string, coveragePercent float64) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer CloseFile(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		c.addTextWithWords(line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	c.pruneWordsByCoverage(coveragePercent)

	return nil
}

// pruneWordsByCoverage reduces the word list to the most frequent words covering a target percentage.
// For example, with coveragePercent=99, keeps only the most common words that together account
// for 99% of all word occurrences. This reduces memory usage while retaining high-frequency vocabulary.
//
//nolint:unused
func (c *Corpus) pruneWordsByCoverage(coveragePercent float64) {
	if len(c.Words) == 0 {
		return
	}

	// Sort words by frequency
	type wordFreq struct {
		word  string
		count uint64
	}

	words := make([]wordFreq, 0, len(c.Words))
	for word, count := range c.Words {
		words = append(words, wordFreq{word, count})
	}

	sort.Slice(words, func(i, j int) bool {
		return words[i].count > words[j].count
	})

	// Calculate target occurrence count
	targetCount := uint64(float64(c.TotalWordsCount) * coveragePercent / 100.0)

	// Accumulate words until target is reached
	newWords := make(map[string]uint64, len(c.Words)/10)
	var coveredCount uint64
	var keptWords int

	for _, wf := range words {
		if coveredCount >= targetCount {
			break
		}
		newWords[wf.word] = wf.count
		coveredCount += wf.count
		keptWords++
	}

	removedWords := len(c.Words) - keptWords
	removedCount := c.TotalWordsCount - coveredCount

	fmt.Printf("Applied word coverage filtering (%.1f%%):\n", coveragePercent)
	fmt.Printf("  Kept: %d unique words (%d total occurrences, %.2f%% coverage)\n",
		keptWords, coveredCount, float64(coveredCount)/float64(c.TotalWordsCount)*100)
	fmt.Printf("  Removed: %d unique words (%d total occurrences, %.2f%% of corpus)\n\n",
		removedWords, removedCount, float64(removedCount)/float64(c.TotalWordsCount)*100)

	c.Words = newWords
	c.TotalWordsCount = coveredCount
}

// LoadJSON loads a Corpus from the specified JSON file.
func LoadJSON(jsonPath string) (*Corpus, error) {
	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("could not open corpus cache: %w", err)
	}
	defer CloseFile(jsonFile)

	var c Corpus
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&c); err != nil {
		return nil, fmt.Errorf("could not decode corpus cache: %w", err)
	}

	return &c, nil
}

// SaveJSON saves the corpus to a JSON file with pretty-printed formatting.
func (c *Corpus) SaveJSON(jsonPath string) error {
	file, err := os.Create(jsonPath)
	if err != nil {
		return fmt.Errorf("could not create corpus cache file: %w", err)
	}
	defer CloseFile(file)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("could not encode corpus cache: %w", err)
	}
	return nil
}

// TopUnigrams returns the top N most frequent unigrams (single characters).
// If n <= 0 or exceeds the total count, returns all unigrams.
func (c *Corpus) TopUnigrams(n int) []CountPair[Unigram] {
	sorted := SortedMap(c.Unigrams)
	if n > 0 && n < len(sorted) {
		return sorted[:n]
	}
	return sorted
}

// TopBigrams returns the top N most frequent bigrams (2-character sequences).
// If n <= 0 or exceeds the total count, returns all bigrams.
func (c *Corpus) TopBigrams(n int) []CountPair[Bigram] {
	sorted := SortedMap(c.Bigrams)
	if n > 0 && n < len(sorted) {
		return sorted[:n]
	}
	return sorted
}

// TopTrigrams returns the top N most frequent trigrams (3-character sequences).
// If n <= 0 or exceeds the total count, returns all trigrams.
func (c *Corpus) TopTrigrams(n int) []CountPair[Trigram] {
	sorted := SortedMap(c.Trigrams)
	if n > 0 && n < len(sorted) {
		return sorted[:n]
	}
	return sorted
}

// TopSkipgrams returns the top N most frequent skipgrams (1st and 3rd chars of trigrams).
// If n <= 0 or exceeds the total count, returns all skipgrams.
func (c *Corpus) TopSkipgrams(n int) []CountPair[Skipgram] {
	sorted := SortedMap(c.Skipgrams)
	if n > 0 && n < len(sorted) {
		return sorted[:n]
	}
	return sorted
}

// TopConsonantBigrams returns the top N most frequent bigrams containing only consonants,
// along with the total count of all consonant-only bigram occurrences and the number of unique consonant-only bigrams.
// Consonants are defined as letters excluding vowels (a, e, i, o, u).
// If n <= 0 or exceeds the filtered count, returns all consonant-only bigrams.
func (c *Corpus) TopConsonantBigrams(n int) ([]CountPair[Bigram], uint64, int) {
	isVowel := func(r rune) bool {
		r = unicode.ToLower(r)
		return r == 'a' || r == 'e' || r == 'i' || r == 'o' || r == 'u'
	}

	isConsonant := func(r rune) bool {
		return unicode.IsLetter(r) && !isVowel(r)
	}

	// Filter bigrams directly into a slice
	pairs := make([]CountPair[Bigram], 0, 512)
	var totalCount uint64
	for bigram, count := range c.Bigrams {
		if isConsonant(bigram[0]) && isConsonant(bigram[1]) {
			pairs = append(pairs, CountPair[Bigram]{bigram, count})
			totalCount += count
		}
	}

	// Sort by count descending
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	uniqueCount := len(pairs)
	if n > 0 && n < len(pairs) {
		return pairs[:n], totalCount, uniqueCount
	}
	return pairs, totalCount, uniqueCount
}

// TopDoubleLetters returns the top N most frequent bigrams where both runes are equal,
// along with the total count of all double-letter bigram occurrences and the number of unique double-letter bigrams.
// If n <= 0 or exceeds the total count, returns all double-letter bigrams.
func (c *Corpus) TopDoubleLetters(n int) ([]CountPair[Bigram], uint64, int) {
	// Filter bigrams directly into a slice
	pairs := make([]CountPair[Bigram], 0, 64)
	var totalCount uint64
	for bigram, count := range c.Bigrams {
		if bigram[0] == bigram[1] {
			pairs = append(pairs, CountPair[Bigram]{bigram, count})
			totalCount += count
		}
	}

	// Sort by count descending
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	uniqueCount := len(pairs)
	if n > 0 && n < len(pairs) {
		return pairs[:n], totalCount, uniqueCount
	}
	return pairs, totalCount, uniqueCount
}

// TopWords returns the top N most frequent words.
// If n <= 0 or exceeds the total count, returns all words.
func (c *Corpus) TopWords(n int) []CountPair[string] {
	sorted := SortedMap(c.Words)
	if n > 0 && n < len(sorted) {
		return sorted[:n]
	}
	return sorted
}

// CorpusInput encapsulates parameters for corpus display.
type CorpusInput struct {
	Corpus *Corpus
	NRows  int
}

// CorpusResult contains corpus statistics ready for display.
type CorpusResult struct {
	Corpus *Corpus
	NRows  int
}

// DisplayCorpus performs pure computation for corpus display.
// This is a simple pass-through since corpus statistics are already computed.
func DisplayCorpus(input CorpusInput) (*CorpusResult, error) {
	return &CorpusResult{
		Corpus: input.Corpus,
		NRows:  input.NRows,
	}, nil
}
