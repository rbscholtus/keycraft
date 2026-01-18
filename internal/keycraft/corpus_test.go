package keycraft

import (
	"testing"
)

func TestAddTextWithWords_Apostrophes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]uint64
	}{
		{
			name:  "contractions with apostrophes",
			input: "don't she'll we'd you're can't won't I'm",
			expected: map[string]uint64{
				"don't":  1,
				"she'll": 1,
				"we'd":   1,
				"you're": 1,
				"can't":  1,
				"won't":  1,
				"i'm":    1,
			},
		},
		{
			name:  "possessives",
			input: "John's book Sarah's car it's working",
			expected: map[string]uint64{
				"john's":  1,
				"book":    1,
				"sarah's": 1,
				"car":     1,
				"it's":    1,
				"working": 1,
			},
		},
		{
			name:  "possessive plural",
			input: "the users' accounts the teachers' lounge",
			expected: map[string]uint64{
				"the":       2,
				"users'":    1,
				"accounts":  1,
				"teachers'": 1,
				"lounge":    1,
			},
		},
		{
			name:  "leading apostrophes trimmed",
			input: "'hello 'world",
			expected: map[string]uint64{
				"hello": 1,
				"world": 1,
			},
		},
		{
			name:  "multiple trailing apostrophes trimmed",
			input: "hello'' world'''",
			expected: map[string]uint64{
				"hello'": 1,
				"world'": 1,
			},
		},
		{
			name:  "archaic contractions",
			input: "'tis 'twas",
			expected: map[string]uint64{
				"tis":  1,
				"twas": 1,
			},
		},
		{
			name:  "unicode apostrophes",
			input: "don't she'll we'd", // Using Unicode right single quotation mark U+2019
			expected: map[string]uint64{
				"don't":  1,
				"she'll": 1,
				"we'd":   1,
			},
		},
		{
			name:  "mixed with regular words",
			input: "I don't think we'll make it there",
			expected: map[string]uint64{
				"i":      1,
				"don't":  1,
				"think":  1,
				"we'll":  1,
				"make":   1,
				"it":     1,
				"there":  1,
			},
		},
		{
			name:  "repeated contractions",
			input: "don't don't don't",
			expected: map[string]uint64{
				"don't": 3,
			},
		},
		{
			name:  "numbers with apostrophes",
			input: "the '90s music",
			expected: map[string]uint64{
				"the":   1,
				"90s":   1,
				"music": 1,
			},
		},
		{
			name:  "empty and whitespace",
			input: "   don't   she'll   ",
			expected: map[string]uint64{
				"don't":  1,
				"she'll": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			corpus := NewCorpus("test")
			corpus.addTextWithWords(tt.input)

			// Check that all expected words are present with correct counts
			for word, expectedCount := range tt.expected {
				actualCount, exists := corpus.Words[word]
				if !exists {
					t.Errorf("Expected word %q not found in corpus", word)
					continue
				}
				if actualCount != expectedCount {
					t.Errorf("Word %q: expected count %d, got %d", word, expectedCount, actualCount)
				}
			}

			// Check that no unexpected words are present
			for word := range corpus.Words {
				if _, expected := tt.expected[word]; !expected {
					t.Errorf("Unexpected word %q found in corpus with count %d", word, corpus.Words[word])
				}
			}

			// Verify total word count
			var expectedTotal uint64
			for _, count := range tt.expected {
				expectedTotal += count
			}
			if corpus.TotalWordsCount != expectedTotal {
				t.Errorf("Expected total word count %d, got %d", expectedTotal, corpus.TotalWordsCount)
			}
		})
	}
}

func TestAddTextWithWords_NGramsUnaffected(t *testing.T) {
	corpus := NewCorpus("test")
	corpus.addTextWithWords("don't worry")

	// Verify that n-grams are still extracted correctly
	// The apostrophe should appear in n-grams since it's part of the text
	if corpus.TotalUnigramsCount == 0 {
		t.Error("Expected unigrams to be extracted")
	}
	if corpus.TotalBigramsCount == 0 {
		t.Error("Expected bigrams to be extracted")
	}

	// Check specific unigrams exist
	if _, exists := corpus.Unigrams[Unigram('d')]; !exists {
		t.Error("Expected unigram 'd' to exist")
	}
	if _, exists := corpus.Unigrams[Unigram('o')]; !exists {
		t.Error("Expected unigram 'o' to exist")
	}
	if _, exists := corpus.Unigrams[Unigram('n')]; !exists {
		t.Error("Expected unigram 'n' to exist")
	}
}

func TestAddTextWithWords_CaseInsensitive(t *testing.T) {
	corpus := NewCorpus("test")
	corpus.addTextWithWords("Don't DON'T don't")

	// All variants should be counted as the same word (lowercase)
	expectedCount := uint64(3)
	actualCount := corpus.Words["don't"]
	if actualCount != expectedCount {
		t.Errorf("Expected count %d for 'don't', got %d", expectedCount, actualCount)
	}

	// Should only have one entry in the map
	if len(corpus.Words) != 1 {
		t.Errorf("Expected 1 unique word, got %d", len(corpus.Words))
	}
}

func TestAddTextWithWords_EmptyInput(t *testing.T) {
	corpus := NewCorpus("test")
	corpus.addTextWithWords("")

	if len(corpus.Words) != 0 {
		t.Errorf("Expected no words for empty input, got %d", len(corpus.Words))
	}
	if corpus.TotalWordsCount != 0 {
		t.Errorf("Expected total word count 0, got %d", corpus.TotalWordsCount)
	}
}

func TestAddTextWithWords_OnlyApostrophes(t *testing.T) {
	corpus := NewCorpus("test")
	corpus.addTextWithWords("''' '' '")

	// Only apostrophes should result in no words
	if len(corpus.Words) != 0 {
		t.Errorf("Expected no words for apostrophe-only input, got %d words: %v", len(corpus.Words), corpus.Words)
	}
}
