package tui

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	kc "github.com/rbscholtus/keycraft/internal/keycraft"
)

// RenderCorpus renders corpus statistics to stdout.
func RenderCorpus(result *kc.CorpusResult) error {
	corpus := result.Corpus
	nrows := result.NRows

	fmt.Printf("Corpus: %s\n\n", corpus.Name)

	fmt.Println(corpusWordLenDistStr(corpus))
	fmt.Println()

	fmt.Println(corpusWordsStr(corpus, nrows))
	fmt.Println()

	fmt.Println(corpusUnigramsStr(corpus, nrows))
	fmt.Println()

	fmt.Println(corpusBigramsStr(corpus, nrows))
	fmt.Println()

	fmt.Println(corpusBigramConsonStr(corpus, nrows))
	fmt.Println()

	fmt.Println(corpusDoubleLettersStr(corpus, nrows))
	fmt.Println()

	fmt.Println(corpusTrigramsStr(corpus, nrows))
	fmt.Println()

	fmt.Println(corpusSkipgramsStr(corpus, nrows))

	return nil
}

// calculatePagination calculates the number of rows per table and number of tables
// for paginated display based on the total number of rows.
// For up to 50 rows of data, the number of rows per table is fixed at 10.
// For over 50 rows of data, the maximum number of tables is fixed at 5.
func calculatePagination(nrows int) (rowsPerTable, numTables int) {
	rowsPerTable, numTables = 10, (nrows+9)/10
	if nrows > 50 {
		rowsPerTable, numTables = (nrows+4)/5, 5
	}
	return rowsPerTable, numTables
}

// renderOuterCorpusTable renders the outer table layout, including pagination and title,
// for displaying corpus statistics.
func renderOuterCorpusTable(inner table.Writer, title string, rowsPerTable int, numTables int) string {
	tables := make(table.Row, 0, numTables)
	p := inner.Pager(table.PageSize(rowsPerTable))
	tables = append(tables, p.Render())
	for p.Location() < numTables {
		next := strings.TrimSpace(p.Next())
		if next == "" {
			break
		}
		tables = append(tables, next)
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

// corpusWordLenDistStr renders the word length distribution as a formatted table.
func corpusWordLenDistStr(corpus *kc.Corpus) string {
	lengthCounts := make(map[int]uint64)

	for word, count := range corpus.Words {
		length := len([]rune(word))
		lengthCounts[length] += count
	}

	t := createSimpleTable()
	t.SetTitle("Word Length Distribution")
	t.SetAutoIndex(false)

	t.AppendHeader(table.Row{"orderby", "Length", "Count", "%", "Cumul%"})

	if len(lengthCounts) > 0 {
		maxLength := slices.Max(slices.Collect(maps.Keys(lengthCounts)))

		cumPct := 0.0
		for length := 1; length <= maxLength; length++ {
			if count, exists := lengthCounts[length]; exists {
				pct := float64(count) / float64(corpus.TotalWordsCount)
				cumPct += pct
				t.AppendRow(table.Row{-length, length, count, pct, cumPct})
			}
		}
	}

	t.AppendFooter(table.Row{"", "Total", corpus.TotalWordsCount, 1.0, ""})

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

	uniqueWords := len(corpus.Words)

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

// corpusUnigramsStr renders the top unigrams as paginated tables.
func corpusUnigramsStr(corpus *kc.Corpus, nrows int) string {
	topUnigrams := corpus.TopUnigrams(nrows)
	rowsPerTable, numTables := calculatePagination(len(topUnigrams))
	t := createSimpleTable()
	t.AppendHeader(table.Row{"orderby", "Ch", "Count", "%", "Cumul%"})
	cumPct := 0.0
	for _, pair := range topUnigrams {
		char := displayChar(rune(pair.Key))
		pct := float64(pair.Count) / float64(corpus.TotalUnigramsCount)
		cumPct += pct
		t.AppendRow(table.Row{pair.Count, char, pair.Count, pct, cumPct})
	}
	title := fmt.Sprintf("Top-%d Unigrams (Total %s occurrences of %s unigrams)",
		len(topUnigrams), Comma(corpus.TotalUnigramsCount), Comma(len(corpus.Unigrams)))
	return renderOuterCorpusTable(t, title, rowsPerTable, numTables)
}

// corpusBigramsStr renders the top bigrams as paginated tables.
func corpusBigramsStr(corpus *kc.Corpus, nrows int) string {
	topBigrams := corpus.TopBigrams(nrows)
	rowsPerTable, numTables := calculatePagination(len(topBigrams))
	t := createSimpleTable()
	t.AppendHeader(table.Row{"orderby", "Bi", "Count", "%", "Cumul%"})
	cumPct := 0.0
	for _, pair := range topBigrams {
		pct := float64(pair.Count) / float64(corpus.TotalBigramsCount)
		cumPct += pct
		t.AppendRow(table.Row{pair.Count, pair.Key.String(), pair.Count, pct, cumPct})
	}
	title := fmt.Sprintf("Top-%d Bigrams (Total %s occurrences of %s bigrams)",
		len(topBigrams), Comma(corpus.TotalBigramsCount), Comma(len(corpus.Bigrams)))
	return renderOuterCorpusTable(t, title, rowsPerTable, numTables)
}

// corpusBigramConsonStr renders the top consonant-only bigrams as paginated tables.
func corpusBigramConsonStr(corpus *kc.Corpus, nrows int) string {
	topConsonantBigrams, totalConsOnly, uniqueConsOnly := corpus.TopConsonantBigrams(nrows)
	rowsPerTable, numTables := calculatePagination(len(topConsonantBigrams))

	t := createSimpleTable()
	t.AppendHeader(table.Row{"orderby", "Bi", "Count", "%", "Cumul%"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "orderby", Hidden: true},
		{Name: "Count", Transformer: Thousands},
		{Name: "%", Transformer: Percentage3},
		{Name: "Cumul%", Transformer: Percentage3},
	})
	cumPct := 0.0
	for _, pair := range topConsonantBigrams {
		pct := float64(pair.Count) / float64(corpus.TotalBigramsCount)
		cumPct += pct
		t.AppendRow(table.Row{pair.Count, pair.Key.String(), pair.Count, pct, cumPct})
	}
	title := fmt.Sprintf("Top-%d Consonant-Only Bigrams (Total %s occurrences of %s bigrams)",
		len(topConsonantBigrams), Comma(totalConsOnly), Comma(uniqueConsOnly))
	return renderOuterCorpusTable(t, title, rowsPerTable, numTables)
}

// corpusDoubleLettersStr renders the top double-letter bigrams as paginated tables.
func corpusDoubleLettersStr(corpus *kc.Corpus, nrows int) string {
	topDoubleLetters, totalDouble, uniqueDouble := corpus.TopDoubleLetters(nrows)
	rowsPerTable, numTables := calculatePagination(len(topDoubleLetters))

	t := createSimpleTable()
	t.AppendHeader(table.Row{"orderby", "Bi", "Count", "%", "Cumul%"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "orderby", Hidden: true},
		{Name: "Count", Transformer: Thousands},
		{Name: "%", Transformer: Percentage3},
		{Name: "Cumul%", Transformer: Percentage3},
	})

	cumPct := 0.0
	for _, pair := range topDoubleLetters {
		pct := float64(pair.Count) / float64(corpus.TotalBigramsCount)
		cumPct += pct
		t.AppendRow(table.Row{pair.Count, pair.Key.String(), pair.Count, pct, cumPct})
	}
	title := fmt.Sprintf("Top-%d Double-Letter Bigrams (Total %s occurrences of %s bigrams)",
		len(topDoubleLetters), Comma(totalDouble), Comma(uniqueDouble))
	return renderOuterCorpusTable(t, title, rowsPerTable, numTables)
}

// corpusTrigramsStr renders the top trigrams as paginated tables.
func corpusTrigramsStr(corpus *kc.Corpus, nrows int) string {
	topTrigrams := corpus.TopTrigrams(nrows)
	rowsPerTable, numTables := calculatePagination(len(topTrigrams))
	t := createSimpleTable()
	t.AppendHeader(table.Row{"orderby", "Tri", "Count", "%", "Cumul%"})
	cumPct := 0.0
	for _, pair := range topTrigrams {
		pct := float64(pair.Count) / float64(corpus.TotalTrigramsCount)
		cumPct += pct
		t.AppendRow(table.Row{pair.Count, pair.Key.String(), pair.Count, pct, cumPct})
	}
	title := fmt.Sprintf("Top-%d Trigrams (Total %s occurrences of %s trigrams)",
		len(topTrigrams), Comma(corpus.TotalTrigramsCount), Comma(len(corpus.Trigrams)))
	return renderOuterCorpusTable(t, title, rowsPerTable, numTables)
}

// corpusSkipgramsStr renders the top skipgrams as paginated tables.
func corpusSkipgramsStr(corpus *kc.Corpus, nrows int) string {
	topSkipgrams := corpus.TopSkipgrams(nrows)
	rowsPerTable, numTables := calculatePagination(len(topSkipgrams))
	t := createSimpleTable()
	t.AppendHeader(table.Row{"orderby", "Skp", "Count", "%", "Cumul%"})
	cumPct := 0.0
	for _, pair := range topSkipgrams {
		pct := float64(pair.Count) / float64(corpus.TotalSkipgramsCount)
		cumPct += pct
		t.AppendRow(table.Row{pair.Count, pair.Key.String(), pair.Count, pct, cumPct})
	}
	title := fmt.Sprintf("Top-%d Skipgrams (Total %s occurrences of %s skipgrams)",
		len(topSkipgrams), Comma(corpus.TotalSkipgramsCount), Comma(len(corpus.Skipgrams)))
	return renderOuterCorpusTable(t, title, rowsPerTable, numTables)
}

// corpusWordsStr renders the top words as a paginated table.
func corpusWordsStr(corpus *kc.Corpus, nrows int) string {
	topWords := corpus.TopWords(nrows)
	rowsPerTable, numTables := calculatePagination(len(topWords))
	t := createSimpleTable()
	t.AppendHeader(table.Row{"orderby", "Word", "Count", "%", "Cumul%"})
	cumPct := 0.0
	for _, pair := range topWords {
		pct := float64(pair.Count) / float64(corpus.TotalWordsCount)
		cumPct += pct
		t.AppendRow(table.Row{pair.Count, pair.Key, pair.Count, pct, cumPct})
	}
	title := fmt.Sprintf("Top-%d Words (Total %s occurrences of %s words)",
		len(topWords), Comma(corpus.TotalWordsCount), Comma(len(corpus.Words)))
	return renderOuterCorpusTable(t, title, rowsPerTable, numTables)
}
