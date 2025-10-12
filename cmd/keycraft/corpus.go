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

// corpusCommand defines the CLI command for displaying corpus statistics.
var corpusCommand = &cli.Command{
	Name:    "corpus",
	Aliases: []string{"c"},
	Usage:   "Display statistics for a text corpus",
	Flags:   flagsSlice("corpus", "rows", "coverage-threshold"),
	Action:  corpusAction,
}

// corpusAction loads the specified corpus and displays its statistics.
// It returns an error if the corpus cannot be loaded.
func corpusAction(c *cli.Context) error {
	corpus, err := getCorpusFromFlags(c)
	if err != nil {
		return err
	}

	nrows := c.Int("rows")

	fmt.Printf("Corpus: %s\n\n", corpus.Name)

	fmt.Println(corpusWordLengthString(corpus))
	fmt.Println()

	fmt.Println(corpusUnigramsString(corpus, nrows))
	fmt.Println()

	fmt.Println(corpusBigramsString(corpus, nrows))
	fmt.Println()

	fmt.Println(corpusTrigramsString(corpus, nrows))
	fmt.Println()

	fmt.Println(corpusSkipgramsString(corpus, nrows))
	fmt.Println()

	fmt.Println(corpusWordsString(corpus, nrows))

	return nil
}

// renderOuterCorpusTable renders the outer table layout, including pagination and title,
// for displaying corpus statistics.
func renderOuterCorpusTable(inner table.Writer, title string, rowsPerTable int, numTables int) string {
	tables := make(table.Row, 0, numTables)
	p := inner.Pager(table.PageSize(rowsPerTable))
	tables = append(tables, p.Render())
	for p.Location() < numTables {
		tables = append(tables, strings.TrimSpace(p.Next()))
	}

	outer := table.NewWriter()
	outer.SetStyle(table.Style{
		Box:     table.BoxStyle{MiddleVertical: " ", MiddleHorizontal: ""},
		Options: table.OptionsNoBorders,
	})
	outer.Style().Title.Align = text.AlignCenter
	outer.SetTitle(title)
	outer.AppendRow(tables)

	return outer.Render()
}

// corpusWordLengthString renders the word length distribution as a formatted table.
func corpusWordLengthString(corpus *kc.Corpus) string {
	lengthCounts := make(map[int]uint64)

	for word, count := range corpus.Words {
		length := len([]rune(word)) // Use runes to count characters properly
		lengthCounts[length] += count
	}

	t := createSimpleTable()
	t.SetTitle("Word Length Distribution")

	t.AppendHeader(table.Row{"orderby", "Length", "Count", "%"})

	maxLength := 0
	for length := range lengthCounts {
		if length > maxLength {
			maxLength = length
		}
	}

	for length := 1; length <= maxLength; length++ {
		if count, exists := lengthCounts[length]; exists {
			pct := float64(count) / float64(corpus.TotalWordsCount)
			t.AppendRow(table.Row{100 - length, fmt.Sprintf("%d chars", length), count, pct})
		}
	}

	t.AppendFooter(table.Row{"", "Total", corpus.TotalWordsCount, 1.0})

	return t.Render()
}

// corpusWordFrequencyString renders the word frequency distribution as a formatted table.
// It shows how many unique words occur N times (e.g., "100 words appear exactly 5 times").
//
//nolint:unused // reserved for future corpus analysis features
func corpusWordFrequencyString(corpus *kc.Corpus) string {
	freqCounts := make(map[uint64]uint64)

	for _, count := range corpus.Words {
		freqCounts[count]++
	}

	t := createSimpleTable()
	t.SetTitle("Word Frequency Distribution")

	t.AppendHeader(table.Row{"orderby", "Occurrences", "# of Words", "%"})

	type freqPair struct {
		freq  uint64
		count uint64
	}
	var pairs []freqPair
	for freq, count := range freqCounts {
		pairs = append(pairs, freqPair{freq, count})
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].freq < pairs[j].freq
		}
		return pairs[i].count > pairs[j].count
	})

	uniqueWords := uint64(len(corpus.Words))

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

// displayChar returns a printable string representation of a given rune.
// Special characters like space, tab, newline, and carriage return are
// represented by symbols (e.g., '_', '\t', '\n', '\r'). Non-printable
// characters are shown as Unicode code points (e.g., 'U+004X').
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

// corpusUnigramsString renders the top unigrams as paginated tables.
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

// corpusBigramsString renders the top bigrams as paginated tables.
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

// corpusTrigramsString renders the top trigrams as paginated tables.
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

// corpusSkipgramsString renders the top skipgrams as paginated tables.
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

// corpusWordsString renders the top words as a paginated table.
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
