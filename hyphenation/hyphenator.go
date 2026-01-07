package hyphenation

// Hyphenator provides word hyphenation using basic rules.
// This is a simplified implementation; a full TeX-style implementation
// would use Liang's pattern-based algorithm.
type Hyphenator struct {
	minWordLen   int // Minimum word length to hyphenate
	minPrefix    int // Minimum characters before hyphen
	minSuffix    int // Minimum characters after hyphen
}

// New creates a new Hyphenator with default settings.
func New() *Hyphenator {
	return &Hyphenator{
		minWordLen: 5, // Don't hyphenate words shorter than 5 chars
		minPrefix:  2, // At least 2 chars before hyphen
		minSuffix:  3, // At least 3 chars after hyphen
	}
}

// Hyphenate returns potential hyphenation points for a word.
// Returns indices where hyphens can be inserted (between characters).
// For example, "typography" might return [2, 5, 7] for "ty-pog-ra-phy".
func (h *Hyphenator) Hyphenate(word string) []int {
	runes := []rune(word)
	n := len(runes)

	// Too short to hyphenate
	if n < h.minWordLen {
		return nil
	}

	var points []int

	// Use simple syllable-based rules for English
	// This is a heuristic approach - not as accurate as TeX patterns
	for i := h.minPrefix; i <= n-h.minSuffix; i++ {
		if h.canBreakAt(runes, i) {
			points = append(points, i)
		}
	}

	return points
}

// canBreakAt checks if we can break at position i (between runes[i-1] and runes[i]).
func (h *Hyphenator) canBreakAt(runes []rune, i int) bool {
	if i < 1 || i >= len(runes) {
		return false
	}

	prev := runes[i-1]
	curr := runes[i]
	n := len(runes)

	// Rule 1: Break between two consonants (if they're different)
	if isConsonant(prev) && isConsonant(curr) && prev != curr {
		// But not after certain consonant clusters that should stay together
		if !isConsonantCluster(prev, curr) {
			return true
		}
	}

	// Rule 2: Break after a vowel followed by a consonant, if next is also consonant
	if i >= 2 && isVowel(runes[i-2]) && isConsonant(prev) && isConsonant(curr) {
		return true
	}

	// Rule 3: Break between vowel and consonant in VCV pattern
	// (prefer breaking before the consonant: "ty-po" not "typ-o")
	if isVowel(prev) && isConsonant(curr) && i+1 < n && isVowel(runes[i+1]) {
		return true
	}

	// Rule 4: Break after common prefixes
	prefix := string(runes[:i])
	if isCommonPrefix(prefix) {
		return true
	}

	// Rule 5: Break before common suffixes
	suffix := string(runes[i:])
	if isCommonSuffix(suffix) {
		return true
	}

	// Rule 6: For long words (>8 chars), allow break after any vowel-consonant pair
	// This provides fallback opportunities for complex words
	if n > 8 && isVowel(prev) && isConsonant(curr) {
		return true
	}

	// Rule 7: Break between two vowels in hiatus (e.g., "cre-ate", "po-em")
	if isVowel(prev) && isVowel(curr) && prev != curr {
		return true
	}

	return false
}

func isVowel(r rune) bool {
	switch r {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U', 'y', 'Y':
		return true
	}
	return false
}

func isConsonant(r rune) bool {
	if r >= 'a' && r <= 'z' {
		return !isVowel(r)
	}
	if r >= 'A' && r <= 'Z' {
		return !isVowel(r)
	}
	return false
}

// isConsonantCluster returns true for consonant pairs that shouldn't be split.
func isConsonantCluster(c1, c2 rune) bool {
	// Common clusters that should stay together
	clusters := map[string]bool{
		"bl": true, "br": true, "ch": true, "ck": true, "cl": true,
		"cr": true, "dr": true, "fl": true, "fr": true, "gh": true,
		"gl": true, "gr": true, "ph": true, "pl": true, "pr": true,
		"sc": true, "sh": true, "sk": true, "sl": true, "sm": true,
		"sn": true, "sp": true, "st": true, "sw": true, "th": true,
		"tr": true, "tw": true, "wh": true, "wr": true,
	}
	pair := string([]rune{c1, c2})
	return clusters[pair] || clusters[string([]rune{c1 + 32, c2 + 32})] // Check lowercase
}

func isCommonPrefix(s string) bool {
	prefixes := map[string]bool{
		"un": true, "re": true, "in": true, "dis": true, "en": true,
		"non": true, "pre": true, "mis": true, "over": true, "sub": true,
		"anti": true, "auto": true, "semi": true, "super": true,
		"trans": true, "inter": true, "under": true, "out": true,
		"con": true, "com": true, "pro": true, "per": true, "ex": true,
		"de": true, "be": true, "fore": true, "counter": true,
	}
	return prefixes[s]
}

func isCommonSuffix(s string) bool {
	suffixes := map[string]bool{
		"ing": true, "tion": true, "sion": true, "ment": true, "ness": true,
		"able": true, "ible": true, "ful": true, "less": true, "ous": true,
		"ive": true, "ly": true, "er": true, "est": true, "al": true,
		"ity": true, "ty": true, "ry": true, "phy": true, "gy": true,
		"ers": true, "ies": true, "ed": true, "es": true, "en": true,
		"ian": true, "ical": true, "ally": true,
		"ence": true, "ance": true, "eous": true, "ious": true,
		"tial": true, "cial": true, "tive": true, "sive": true,
		"phers": true, "ther": true,
		"graph": true, "raphy": true, "logy": true, "sophy": true,
	}
	return suffixes[s]
}
