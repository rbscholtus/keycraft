package keycraft

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// GenerateInput captures CLI inputs for generation.
type GenerateInput struct {
	ConfigPath      string // resolved .gen file path
	MaxLayouts      int    // from --max-layouts flag (default 1500, 0=all)
	Seed            uint64 // from --seed flag (0=timestamp)
	Optimize        bool   // from --optimize flag
	KeepUnoptimized bool   // from --keep-unoptimized flag
}

// PositionType defines what kind of allocation should happen at a position.
type PositionType uint8

const (
	PositionUnused PositionType = iota // ~ (not allocated)
	PositionFixed                      // specific char or _
	PositionRandom                     // 0 (random from charset)
	PositionGroup                      // 1-9 (from group set)
)

// PositionSpec defines what should go in a position.
type PositionSpec struct {
	Type      PositionType
	FixedChar rune // used when Type == PositionFixed
	GroupNum  int  // used when Type == PositionGroup (1-9)
}

// GenerationConfig represents a parsed .gen file.
type GenerationConfig struct {
	LayoutType LayoutType          // ROWSTAG, ANGLEMOD, ORTHO, or COLSTAG
	Template   [42]PositionSpec    // Specification for each position
	Charset    []rune              // All characters to allocate
	Groups     map[int][]rune      // Group number -> character set
	FilePath   string              // Original file path for error messages
	LineNums   map[string]int      // Line numbers for error messages (key -> line number)
}

// GenerationResult holds the results of a generation run.
type GenerationResult struct {
	Layouts      []*SplitLayout // Generated layouts
	LayoutPaths  []string       // Paths where layouts were saved
	TotalPerms   int            // Total permutations computed
	Generated    int            // Number actually generated (may be limited by MaxLayouts)
	Warnings     []string       // Any warnings during generation
}

// ParseConfigFile parses a .gen file and returns a GenerationConfig.
func ParseConfigFile(path string) (*GenerationConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %w", err)
	}
	defer CloseFile(file)

	config := &GenerationConfig{
		Groups:   make(map[int][]rune),
		FilePath: path,
		LineNums: make(map[string]int),
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	templateLines := make([]string, 0, 4)
	inTemplate := false
	templateStartLine := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// First non-empty, non-comment line is layout type
		if config.LayoutType == 0 && !inTemplate && len(templateLines) == 0 {
			lt, ok := parseLayoutType(trimmed)
			if !ok {
				return nil, fmt.Errorf("line %d: invalid layout type %q (must be rowstag, anglemod, ortho, or colstag)", lineNum, trimmed)
			}
			config.LayoutType = lt
			config.LineNums["layoutType"] = lineNum
			inTemplate = true
			templateStartLine = lineNum + 1
			continue
		}

		// Collect template lines (4 lines)
		if inTemplate && len(templateLines) < 4 {
			templateLines = append(templateLines, line) // Keep original spacing
			if len(templateLines) == 4 {
				inTemplate = false
			}
			continue
		}

		// Parse key=value definitions
		if strings.Contains(trimmed, "=") {
			key, value, found := strings.Cut(trimmed, "=")
			if !found {
				return nil, fmt.Errorf("line %d: malformed definition (expected key=value)", lineNum)
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)

			if key == "charset" {
				config.Charset = parseCharsetValue(value)
				config.LineNums["charset"] = lineNum
			} else if len(key) >= 3 && key[:3] == "set" {
				// Parse setN where N is 1-9
				if len(key) != 4 {
					return nil, fmt.Errorf("line %d: invalid set name %q (expected set1 through set9)", lineNum, key)
				}
				n := int(key[3] - '0')
				if n < 1 || n > 9 {
					return nil, fmt.Errorf("line %d: invalid set number %q (must be 1-9)", lineNum, key)
				}
				config.Groups[n] = parseCharsetValue(value)
				config.LineNums[key] = lineNum
			} else {
				return nil, fmt.Errorf("line %d: unknown definition %q", lineNum, key)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse template lines
	if len(templateLines) != 4 {
		return nil, fmt.Errorf("expected 4 template lines, got %d", len(templateLines))
	}

	if err := parseTemplate(config, templateLines, templateStartLine); err != nil {
		return nil, err
	}

	return config, nil
}

// ParseConfigString parses a .gen config from a string and returns a GenerationConfig.
// This is useful for testing with embedded config strings.
func ParseConfigString(content string) (*GenerationConfig, error) {
	config := &GenerationConfig{
		Groups:   make(map[int][]rune),
		FilePath: "<string>",
		LineNums: make(map[string]int),
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	templateLines := make([]string, 0, 4)
	inTemplate := false
	templateStartLine := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// First non-empty, non-comment line is layout type
		if config.LayoutType == 0 && !inTemplate && len(templateLines) == 0 {
			lt, ok := parseLayoutType(trimmed)
			if !ok {
				return nil, fmt.Errorf("line %d: invalid layout type %q (must be rowstag, anglemod, ortho, or colstag)", lineNum, trimmed)
			}
			config.LayoutType = lt
			config.LineNums["layoutType"] = lineNum
			inTemplate = true
			templateStartLine = lineNum + 1
			continue
		}

		// Collect template lines (4 lines)
		if inTemplate && len(templateLines) < 4 {
			templateLines = append(templateLines, line) // Keep original spacing
			if len(templateLines) == 4 {
				inTemplate = false
			}
			continue
		}

		// Parse key=value definitions
		if strings.Contains(trimmed, "=") {
			key, value, found := strings.Cut(trimmed, "=")
			if !found {
				return nil, fmt.Errorf("line %d: malformed definition (expected key=value)", lineNum)
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)

			if key == "charset" {
				config.Charset = parseCharsetValue(value)
				config.LineNums["charset"] = lineNum
			} else if len(key) >= 3 && key[:3] == "set" {
				// Parse setN where N is 1-9
				if len(key) != 4 {
					return nil, fmt.Errorf("line %d: invalid set name %q (expected set1 through set9)", lineNum, key)
				}
				n := int(key[3] - '0')
				if n < 1 || n > 9 {
					return nil, fmt.Errorf("line %d: invalid set number %q (must be 1-9)", lineNum, key)
				}
				config.Groups[n] = parseCharsetValue(value)
				config.LineNums[key] = lineNum
			} else {
				return nil, fmt.Errorf("line %d: unknown definition %q", lineNum, key)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	// Parse template lines
	if len(templateLines) != 4 {
		return nil, fmt.Errorf("expected 4 template lines, got %d", len(templateLines))
	}

	if err := parseTemplate(config, templateLines, templateStartLine); err != nil {
		return nil, err
	}

	return config, nil
}

// parseLayoutType converts a string to a LayoutType.
func parseLayoutType(s string) (LayoutType, bool) {
	switch strings.ToLower(s) {
	case "rowstag":
		return ROWSTAG, true
	case "anglemod":
		return ANGLEMOD, true
	case "ortho":
		return ORTHO, true
	case "colstag":
		return COLSTAG, true
	default:
		return 0, false
	}
}

// parseCharsetValue parses a charset or set value, handling _ for space.
func parseCharsetValue(value string) []rune {
	runes := make([]rune, 0, len(value))
	for _, r := range value {
		if r == '_' {
			runes = append(runes, ' ')
		} else {
			runes = append(runes, r)
		}
	}
	return runes
}

// parseTemplate parses the 4 template lines into the config.Template array.
func parseTemplate(config *GenerationConfig, lines []string, startLine int) error {
	posIdx := 0

	for rowIdx, line := range lines {
		lineNum := startLine + rowIdx
		tokens := tokenizeLine(line)

		expectedTokens := 12
		if rowIdx == 3 {
			expectedTokens = 6 // thumb row
		}

		if len(tokens) != expectedTokens {
			return fmt.Errorf("line %d: expected %d positions, got %d", lineNum, expectedTokens, len(tokens))
		}

		for _, token := range tokens {
			spec, err := parsePositionToken(token)
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNum, err)
			}
			config.Template[posIdx] = spec
			posIdx++
		}
	}

	if posIdx != 42 {
		return fmt.Errorf("expected 42 positions, got %d", posIdx)
	}

	return nil
}

// tokenizeLine splits a template line into tokens, handling whitespace.
func tokenizeLine(line string) []string {
	return strings.Fields(line)
}

// parsePositionToken parses a single position token into a PositionSpec.
func parsePositionToken(token string) (PositionSpec, error) {
	if len(token) == 0 {
		return PositionSpec{}, fmt.Errorf("empty token")
	}

	// Handle special tokens
	switch token {
	case "~":
		return PositionSpec{Type: PositionUnused}, nil
	case "_":
		return PositionSpec{Type: PositionFixed, FixedChar: ' '}, nil
	case "0":
		return PositionSpec{Type: PositionRandom}, nil
	}

	// Handle group numbers 1-9
	if len(token) == 1 && token[0] >= '1' && token[0] <= '9' {
		return PositionSpec{Type: PositionGroup, GroupNum: int(token[0] - '0')}, nil
	}

	// Handle multi-character tokens (escape sequences)
	if len(token) == 2 {
		switch token {
		case "~~":
			return PositionSpec{Type: PositionFixed, FixedChar: '~'}, nil
		case "__":
			return PositionSpec{Type: PositionFixed, FixedChar: '_'}, nil
		case "##":
			return PositionSpec{Type: PositionFixed, FixedChar: '#'}, nil
		}
	}

	// Single character = fixed character
	runes := []rune(token)
	if len(runes) == 1 {
		return PositionSpec{Type: PositionFixed, FixedChar: runes[0]}, nil
	}

	return PositionSpec{}, fmt.Errorf("invalid position token %q", token)
}

// ValidateConfig validates a GenerationConfig and returns any errors.
func ValidateConfig(config *GenerationConfig) error {
	var errors []string

	// Layout type validation (already done in parsing, but double-check)
	if config.LayoutType < ROWSTAG || config.LayoutType > COLSTAG {
		errors = append(errors, "invalid layout type")
	}

	// Charset validation
	if len(config.Charset) == 0 {
		errors = append(errors, fmt.Sprintf("line %d: charset must be defined and non-empty", config.LineNums["charset"]))
	}

	// Check for duplicates in charset
	charsetSet := make(map[rune]bool)
	for _, r := range config.Charset {
		if charsetSet[r] {
			errors = append(errors, fmt.Sprintf("duplicate character %q in charset", r))
		}
		charsetSet[r] = true
	}

	// Count positions by type and track fixed characters
	fixedChars := make(map[rune]int)
	groupPositions := make(map[int]int) // group number -> count of positions
	allocatableCount := 0

	for _, spec := range config.Template {
		switch spec.Type {
		case PositionFixed:
			fixedChars[spec.FixedChar]++
			allocatableCount++
		case PositionRandom:
			allocatableCount++
		case PositionGroup:
			groupPositions[spec.GroupNum]++
			allocatableCount++
		case PositionUnused:
			// Not allocatable
		}
	}

	// Check fixed characters appear only once
	for r, count := range fixedChars {
		if count > 1 {
			errors = append(errors, fmt.Sprintf("fixed character %q appears %d times in template", r, count))
		}
	}

	// Check fixed characters are in charset
	for r := range fixedChars {
		if !charsetSet[r] {
			errors = append(errors, fmt.Sprintf("fixed character %q not in charset", r))
		}
	}

	// Group validation
	for groupNum, posCount := range groupPositions {
		groupChars, exists := config.Groups[groupNum]
		if !exists {
			errors = append(errors, fmt.Sprintf("group %d used in template but set%d not defined", groupNum, groupNum))
			continue
		}

		// Check group has enough characters
		if len(groupChars) < posCount {
			errors = append(errors, fmt.Sprintf("set%d has %d characters but %d positions need it", groupNum, len(groupChars), posCount))
		}

		// Check for duplicates in group
		groupSet := make(map[rune]bool)
		for _, r := range groupChars {
			if groupSet[r] {
				errors = append(errors, fmt.Sprintf("duplicate character %q in set%d", r, groupNum))
			}
			groupSet[r] = true
		}

		// Check group is subset of charset
		for _, r := range groupChars {
			if !charsetSet[r] {
				errors = append(errors, fmt.Sprintf("character %q in set%d not in charset", r, groupNum))
			}
		}
	}

	// Warn if setN is defined but not used (this is just a warning, not an error)
	for groupNum := range config.Groups {
		if groupPositions[groupNum] == 0 {
			// Could add to warnings instead of errors
			errors = append(errors, fmt.Sprintf("set%d defined but not used in template", groupNum))
		}
	}

	// Character count validation
	if len(config.Charset) != allocatableCount {
		errors = append(errors, fmt.Sprintf("charset has %d characters but %d allocatable positions", len(config.Charset), allocatableCount))
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// GeneratePermutations generates all valid permutation combinations for the groups.
// When groups share characters, this uses constrained generation to ensure no
// character appears in multiple groups within a single permutation.
// Returns a slice of permutation sets, where each set maps group number to the permutation.
func GeneratePermutations(config *GenerationConfig, maxLayouts int) ([]map[int][]rune, int, []string) {
	var warnings []string

	// Get groups used in template (sorted for determinism)
	groupNums := make([]int, 0, len(config.Groups))
	groupPositions := make(map[int]int)
	for _, spec := range config.Template {
		if spec.Type == PositionGroup {
			groupPositions[spec.GroupNum]++
		}
	}
	for g := range groupPositions {
		groupNums = append(groupNums, g)
	}
	slices.Sort(groupNums)

	// If no groups, return single empty permutation
	if len(groupNums) == 0 {
		return []map[int][]rune{{}}, 1, warnings
	}

	// Check for overlapping groups and warn
	if hasOverlappingGroups(config, groupNums) {
		warnings = append(warnings, "groups share characters; some permutations will be skipped to avoid duplicates")
	}

	// Count total valid permutations first (without allocating memory for all of them)
	// Use string instead of map for better performance (immutable, passed by value)
	totalPerms := countConstrainedPerms(config, groupNums, groupPositions, 0, "")

	if totalPerms > 1000 {
		warnings = append(warnings, fmt.Sprintf("generating %d permutations (this may take a while)", totalPerms))
	}

	// Determine limit for generation
	limit := 0 // 0 means no limit
	if maxLayouts > 0 && totalPerms > maxLayouts {
		warnings = append(warnings, fmt.Sprintf("limiting to %d layouts (total valid permutations: %d)", maxLayouts, totalPerms))
		limit = maxLayouts
	}

	// Generate constrained permutations (respecting character uniqueness across groups)
	result := generateConstrainedPerms(config, groupNums, groupPositions, 0, "", limit)

	return result, totalPerms, warnings
}

// hasOverlappingGroups checks if any character appears in multiple groups.
func hasOverlappingGroups(config *GenerationConfig, groupNums []int) bool {
	seen := make(map[rune]bool)
	for _, groupNum := range groupNums {
		for _, c := range config.Groups[groupNum] {
			if seen[c] {
				return true
			}
			seen[c] = true
		}
	}
	return false
}

// countConstrainedPerms counts the total number of valid permutations without
// allocating memory for them. This mirrors generateConstrainedPerms logic.
// Uses string instead of map for better performance (immutable, passed by value).
func countConstrainedPerms(
	config *GenerationConfig,
	groupNums []int,
	groupPositions map[int]int,
	idx int,
	usedChars string,
) int {
	// Base case: all groups processed
	if idx == len(groupNums) {
		return 1
	}

	groupNum := groupNums[idx]
	k := groupPositions[groupNum]

	// Filter out already-used characters from this group's available chars
	availableChars := make([]rune, 0, len(config.Groups[groupNum]))
	for _, c := range config.Groups[groupNum] {
		if !strings.ContainsRune(usedChars, c) {
			availableChars = append(availableChars, c)
		}
	}

	// Not enough characters available for this group
	n := len(availableChars)
	if n < k {
		return 0
	}

	// Generate all ordered selections for this group and count recursively
	groupPerms := generateOrderedSelections(availableChars, k)
	total := 0

	for _, perm := range groupPerms {
		// Mark these chars as used for subsequent groups (string concatenation)
		newUsed := usedChars + string(perm)

		// Recursively count permutations for remaining groups
		total += countConstrainedPerms(config, groupNums, groupPositions, idx+1, newUsed)
	}

	return total
}

// generateConstrainedPerms recursively generates permutations, ensuring no character
// is used by multiple groups. Characters used by earlier groups are excluded from
// later groups' available character sets.
// Uses string instead of map for better performance (immutable, passed by value).
func generateConstrainedPerms(
	config *GenerationConfig,
	groupNums []int,
	groupPositions map[int]int,
	idx int,
	usedChars string,
	maxLayouts int,
) []map[int][]rune {
	// Base case: all groups processed
	if idx == len(groupNums) {
		return []map[int][]rune{{}}
	}

	groupNum := groupNums[idx]
	k := groupPositions[groupNum]

	// Filter out already-used characters from this group's available chars
	availableChars := make([]rune, 0, len(config.Groups[groupNum]))
	for _, c := range config.Groups[groupNum] {
		if !strings.ContainsRune(usedChars, c) {
			availableChars = append(availableChars, c)
		}
	}

	// Not enough characters available for this group
	if len(availableChars) < k {
		return nil
	}

	// Generate all ordered selections for this group
	groupPerms := generateOrderedSelections(availableChars, k)
	var result []map[int][]rune

	for _, perm := range groupPerms {
		// Mark these chars as used for subsequent groups (string concatenation)
		newUsed := usedChars + string(perm)

		// Recursively get permutations for remaining groups
		remainingLimit := 0
		if maxLayouts > 0 {
			remainingLimit = maxLayouts - len(result)
			if remainingLimit <= 0 {
				break
			}
		}
		subPerms := generateConstrainedPerms(config, groupNums, groupPositions, idx+1, newUsed, remainingLimit)

		// Combine this group's permutation with all sub-permutations
		for _, subPerm := range subPerms {
			combined := make(map[int][]rune, len(subPerm)+1)
			for g, p := range subPerm {
				combined[g] = p
			}
			combined[groupNum] = perm
			result = append(result, combined)

			if maxLayouts > 0 && len(result) >= maxLayouts {
				return result
			}
		}
	}

	return result
}

// generateOrderedSelections generates all P(n,k) ordered selections of k items from n.
// This is all permutations of k items chosen from the input slice.
func generateOrderedSelections(chars []rune, k int) [][]rune {
	n := len(chars)
	if k > n {
		return nil
	}
	if k == 0 {
		return [][]rune{{}}
	}

	// Calculate P(n,k) = n!/(n-k)!
	count := 1
	for i := 0; i < k; i++ {
		count *= (n - i)
	}

	result := make([][]rune, 0, count)

	// Use iterative algorithm with tracking of used indices
	var generate func(current []rune, used []bool)
	generate = func(current []rune, used []bool) {
		if len(current) == k {
			perm := make([]rune, k)
			copy(perm, current)
			result = append(result, perm)
			return
		}
		for i := 0; i < n; i++ {
			if !used[i] {
				used[i] = true
				current = append(current, chars[i])
				generate(current, used)
				current = current[:len(current)-1]
				used[i] = false
			}
		}
	}

	generate(make([]rune, 0, k), make([]bool, n))
	return result
}

// GenerateLayout generates a single layout from config and a specific permutation.
func GenerateLayout(config *GenerationConfig, groupPerm map[int][]rune, seed uint64, permIndex int) *SplitLayout {
	runes := [42]rune{}

	// Track which characters have been used
	used := make(map[rune]bool)

	// Track group position indices (for assigning permutation chars to positions)
	groupPosIdx := make(map[int]int)

	// Step 1: Allocate fixed characters
	for i, spec := range config.Template {
		if spec.Type == PositionFixed {
			runes[i] = spec.FixedChar
			used[spec.FixedChar] = true
		}
	}

	// Step 2: Allocate group permutations
	for i, spec := range config.Template {
		if spec.Type == PositionGroup {
			perm := groupPerm[spec.GroupNum]
			idx := groupPosIdx[spec.GroupNum]
			runes[i] = perm[idx]
			used[perm[idx]] = true
			groupPosIdx[spec.GroupNum] = idx + 1
		}
	}

	// Step 3: Allocate random positions
	// Collect remaining characters
	remaining := make([]rune, 0, len(config.Charset))
	for _, r := range config.Charset {
		if !used[r] {
			remaining = append(remaining, r)
		}
	}

	// Shuffle remaining characters
	rng := NewLockedRNG(seed+uint64(permIndex), 0)
	ShuffleSlice(rng, remaining)

	// Assign to random positions
	randomIdx := 0
	for i, spec := range config.Template {
		if spec.Type == PositionRandom {
			runes[i] = remaining[randomIdx]
			randomIdx++
		}
	}

	// Generate layout name with permutation index
	name := generateLayoutNameFromRunes(runes, permIndex)

	return NewSplitLayout(name, config.LayoutType, runes)
}

// generateLayoutNameFromRunes creates a layout name from runes and permutation index.
// Format: _<homekeys><thumbkeys>-<4hex>
func generateLayoutNameFromRunes(runes [42]rune, permIndex int) string {
	var b strings.Builder
	b.Grow(20)

	b.WriteByte('_') // Prefix with underscore

	// Home row positions: 13-16, 19-22
	// Thumb row positions: 36-41
	for _, i := range [14]int{13, 14, 15, 16, 19, 20, 21, 22, 36, 37, 38, 39, 40, 41} {
		if r := runes[i]; 'a' <= r && r <= 'z' {
			b.WriteRune(r)
		}
	}

	b.WriteByte('-')

	// 4-digit hex suffix from permutation index
	const hex = "0123456789abcdef"
	suffix := uint16(permIndex & 0xFFFF)
	for i := 3; i >= 0; i-- {
		b.WriteByte(hex[(suffix>>(i*4))&0xF])
	}

	return b.String()
}

// GenerateFromConfig generates all layouts from a config.
func GenerateFromConfig(config *GenerationConfig, input GenerateInput, layoutsDir string) (*GenerationResult, error) {
	result := &GenerationResult{
		Layouts:     make([]*SplitLayout, 0),
		LayoutPaths: make([]string, 0),
	}

	// Generate permutations
	perms, totalPerms, warnings := GeneratePermutations(config, input.MaxLayouts)
	result.TotalPerms = totalPerms
	result.Warnings = warnings

	// Generate each layout
	for i, perm := range perms {
		layout := GenerateLayout(config, perm, input.Seed, i)
		result.Layouts = append(result.Layouts, layout)

		// Save layout
		layoutPath := filepath.Join(layoutsDir, layout.Name+".klf")
		if err := layout.SaveToFile(layoutPath); err != nil {
			return nil, fmt.Errorf("failed to save layout %s: %w", layout.Name, err)
		}
		result.LayoutPaths = append(result.LayoutPaths, layoutPath)
	}

	result.Generated = len(result.Layouts)
	return result, nil
}

// ComputeDefaultPins computes the default pinned keys for a generated layout.
// Default pinning: Pin all non-random positions (unused, fixed, group chars, and space).
func ComputeDefaultPins(config *GenerationConfig, layout *SplitLayout) PinnedKeys {
	var pinned PinnedKeys

	for i, spec := range config.Template {
		switch spec.Type {
		case PositionUnused:
			pinned[i] = true // Always pin unused
		case PositionFixed:
			pinned[i] = true // Pin fixed characters
		case PositionGroup:
			pinned[i] = true // Pin group characters
		case PositionRandom:
			pinned[i] = false // Free to move
		}
	}

	return pinned
}
