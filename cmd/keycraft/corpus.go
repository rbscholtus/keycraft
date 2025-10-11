// corpus.go implements the "corpus" command for viewing corpus statistics
package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

// corpusCommand defines the CLI command for viewing corpus statistics
var corpusCommand = &cli.Command{
	Name:    "corpus",
	Aliases: []string{"c"},
	Usage:   "Display statistics for a text corpus",
	Flags:   flagsSlice("corpus", "rows", "coverage-threshold"),
	Action:  corpusAction,
}

// corpusAction loads the corpus and displays statistics
func corpusAction(c *cli.Context) error {
	corpus, err := getCorpusFromFlags(c)
	if err != nil {
		return err
	}

	nrows := c.Int("rows")

	fmt.Printf("Corpus: %s\n\n", corpus.Name)

	// // Print word frequency distribution
	// fmt.Println(corpusWordFrequencyString(corpus))
	// fmt.Println()

	// Print word length distribution
	fmt.Println(corpusWordLengthString(corpus))
	fmt.Println()

	// Print unigrams
	fmt.Println(corpusUnigramsString(corpus, nrows))
	fmt.Println()

	// Print bigrams
	fmt.Println(corpusBigramsString(corpus, nrows))
	fmt.Println()

	// Print trigrams
	fmt.Println(corpusTrigramsString(corpus, nrows))
	fmt.Println()

	// Print skipgrams
	fmt.Println(corpusSkipgramsString(corpus, nrows))
	fmt.Println()

	// Print words
	fmt.Println(corpusWordsString(corpus, nrows))

	return nil
}

// corpusWordLengthString renders word length distribution as a table
func corpusWordLengthString(corpus *kc.Corpus) string {
	lengthCounts := make(map[int]uint64)

	for word, count := range corpus.Words {
		length := len([]rune(word)) // Use runes to count characters properly
		lengthCounts[length] += count
	}

	t := createSimpleTable()
	t.SetTitle("Word Length Distribution")

	t.AppendHeader(table.Row{"orderby", "Length", "Count", "%"})

	// Find max length to iterate through
	maxLength := 0
	for length := range lengthCounts {
		if length > maxLength {
			maxLength = length
		}
	}

	// Add rows for each length
	for length := 1; length <= maxLength; length++ {
		if count, exists := lengthCounts[length]; exists {
			pct := float64(count) / float64(corpus.TotalWordsCount)
			t.AppendRow(table.Row{100 - length, fmt.Sprintf("%d chars", length), count, pct})
		}
	}

	t.AppendFooter(table.Row{"", "Total", corpus.TotalWordsCount, 1.0})

	return t.Render()
}

// corpusWordFrequencyString renders word frequency distribution as a table
// Shows how many words occur N times (e.g., "100 words appear exactly 5 times")
func corpusWordFrequencyString(corpus *kc.Corpus) string {
	freqCounts := make(map[uint64]uint64)

	// Count how many words have each frequency
	for _, count := range corpus.Words {
		freqCounts[count]++
	}

	t := createSimpleTable()
	t.SetTitle("Word Frequency Distribution")

	t.AppendHeader(table.Row{"orderby", "Occurrences", "# of Words", "%"})

	// Sort frequencies to display in order
	type freqPair struct {
		freq  uint64
		count uint64
	}
	var pairs []freqPair
	for freq, count := range freqCounts {
		pairs = append(pairs, freqPair{freq, count})
	}

	// Sort by count (number of words), descending
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].freq < pairs[j].freq // optional tie-breaker
		}
		return pairs[i].count > pairs[j].count // descending by count
	})

	uniqueWords := uint64(len(corpus.Words))

	// Add rows
	for _, pair := range pairs {
		pct := float64(pair.count) / float64(uniqueWords)
		label := fmt.Sprintf("%d times", pair.freq)
		if pair.freq == 1 {
			label = "1 time"
		}
		t.AppendRow(table.Row{pair.freq, label, pair.count, pct})
	}

	t.AppendFooter(table.Row{"", "Total unique", uniqueWords, 1.0})

	return t.Render()
}

// corpusUnigramsString renders top unigrams as paginated tables
func corpusUnigramsString(corpus *kc.Corpus, nrows int) string {
	rowsPerTable, numTables := 10, (nrows+9)/10
	if nrows > 50 {
		rowsPerTable, numTables = (nrows+4)/5, 5
	}
	topUnigrams := corpus.TopUnigrams(nrows)
	inner := createSimpleTable()
	inner.AppendHeader(table.Row{"orderby", "Char", "Count", "%"})
	for _, pair := range topUnigrams {
		char := displayChar(rune(pair.Key))
		pct := float64(pair.Count) / float64(corpus.TotalUnigramsCount)
		inner.AppendRow(table.Row{pair.Count, char, pair.Count, pct})
	}
	title := fmt.Sprintf("Top Unigrams (Total %s)", Comma(corpus.TotalUnigramsCount))
	return renderOuterCorpusTable(inner, title, rowsPerTable, numTables)
}

// corpusBigramsString renders top bigrams as paginated tables
func corpusBigramsString(corpus *kc.Corpus, nrows int) string {
	rowsPerTable, numTables := 10, (nrows+9)/10
	if nrows > 50 {
		rowsPerTable, numTables = (nrows+4)/5, 5
	}
	topBigrams := corpus.TopBigrams(nrows)
	inner := createSimpleTable()
	inner.AppendHeader(table.Row{"orderby", "Bigram", "Count", "%"})
	for _, pair := range topBigrams {
		pct := float64(pair.Count) / float64(corpus.TotalBigramsCount)
		inner.AppendRow(table.Row{pair.Count, pair.Key.String(), pair.Count, pct})
	}
	title := fmt.Sprintf("Top Bigrams (Total %s)", Comma(corpus.TotalBigramsCount))
	return renderOuterCorpusTable(inner, title, rowsPerTable, numTables)
}

// corpusTrigramsString renders top trigrams as paginated tables
func corpusTrigramsString(corpus *kc.Corpus, nrows int) string {
	rowsPerTable, numTables := 10, (nrows+9)/10
	if nrows > 50 {
		rowsPerTable, numTables = (nrows+4)/5, 5
	}
	topTrigrams := corpus.TopTrigrams(nrows)
	inner := createSimpleTable()
	inner.AppendHeader(table.Row{"orderby", "Trigram", "Count", "%"})
	for _, pair := range topTrigrams {
		pct := float64(pair.Count) / float64(corpus.TotalTrigramsCount)
		inner.AppendRow(table.Row{pair.Count, pair.Key.String(), pair.Count, pct})
	}
	title := fmt.Sprintf("Top Trigrams (Total %s)", Comma(corpus.TotalTrigramsCount))
	return renderOuterCorpusTable(inner, title, rowsPerTable, numTables)
}

// corpusSkipgramsString renders top skipgrams as paginated tables
func corpusSkipgramsString(corpus *kc.Corpus, nrows int) string {
	rowsPerTable, numTables := 10, (nrows+9)/10
	if nrows > 50 {
		rowsPerTable, numTables = (nrows+4)/5, 5
	}
	topSkipgrams := corpus.TopSkipgrams(nrows)
	inner := createSimpleTable()
	inner.AppendHeader(table.Row{"orderby", "Skipgram", "Count", "%"})
	for _, pair := range topSkipgrams {
		pct := float64(pair.Count) / float64(corpus.TotalSkipgramsCount)
		inner.AppendRow(table.Row{pair.Count, pair.Key.String(), pair.Count, pct})
	}
	title := fmt.Sprintf("Top Skipgrams (Total %s)", Comma(corpus.TotalSkipgramsCount))
	return renderOuterCorpusTable(inner, title, rowsPerTable, numTables)
}

// renderOuterCorpusTable renders the outer table layout with pagination and title.
func renderOuterCorpusTable(inner table.Writer, title string, rowsPerTable int, numTables int) string {
	tables := make(table.Row, 0, numTables)
	p := inner.Pager(table.PageSize(rowsPerTable))
	tables = append(tables, p.Render())
	for p.Location() < numTables {
		tables = append(tables, strings.TrimSpace(p.Next()))
	}

	outer := table.NewWriter()
	outer.SetStyle(table.Style{
		Box:     table.BoxStyle{MiddleVertical: " "},
		Options: table.OptionsDefault,
	})
	outer.Style().Title.Align = text.AlignCenter
	outer.SetTitle(title)
	outer.AppendRow(tables)

	return outer.Render()
}

// corpusWordsString renders top words as a table using helper functions.
func corpusWordsString(corpus *kc.Corpus, nrows int) string {
	rowsPerTable, numTables := 10, (nrows+9)/10
	if nrows > 50 {
		rowsPerTable, numTables = (nrows+4)/5, 5
	}
	topWords := corpus.TopWords(nrows)
	inner := createSimpleTable()
	inner.AppendHeader(table.Row{"orderby", "Word", "Count", "%"})
	for _, pair := range topWords {
		pct := float64(pair.Count) / float64(corpus.TotalWordsCount)
		inner.AppendRow(table.Row{pair.Count, pair.Key, pair.Count, pct})
	}
	title := fmt.Sprintf("Top Words (Total %s)", Comma(corpus.TotalWordsCount))
	return renderOuterCorpusTable(inner, title, rowsPerTable, numTables)
}

// displayChar returns a printable representation of a rune
func displayChar(r rune) string {
	switch r {
	case ' ':
		return "_"
	case '\t':
		return "\\t"
	case '\n':
		return "\\n"
	case '\r':
		return "\\r"
	default:
		if unicode.IsPrint(r) {
			return fmt.Sprintf("%c", r)
		}
		return fmt.Sprintf("U+%04X", r)
	}
}
