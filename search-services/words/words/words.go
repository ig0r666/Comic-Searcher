package words

import (
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
)

func IsStopWord(word string) bool {
	switch word {
	case "a", "about", "above", "after", "again", "against", "all", "am", "an",
		"and", "any", "are", "as", "at", "be", "because", "been", "before",
		"being", "below", "between", "both", "but", "by", "can", "did", "do",
		"does", "doing", "don", "down", "during", "each", "few", "for", "from",
		"further", "had", "has", "have", "having", "he", "her", "here", "hers",
		"herself", "him", "himself", "his", "how", "i", "if", "in", "into", "is",
		"it", "its", "itself", "just", "me", "more", "most", "my", "myself",
		"no", "nor", "not", "now", "of", "off", "on", "once", "only", "or",
		"other", "our", "ours", "ourselves", "out", "over", "own", "s", "same",
		"she", "should", "so", "some", "such", "t", "than", "that", "the", "their",
		"theirs", "them", "themselves", "then", "there", "these", "they",
		"this", "those", "through", "to", "too", "under", "until", "up",
		"very", "was", "we", "were", "what", "when", "where", "which", "while",
		"who", "whom", "why", "will", "with", "you", "your", "yours", "yourself",
		"yourselves", "re", "ve", "d", "ll", "m":
		return true
	}
	return false
}

func splitIntoWords(input string) []string {
	return strings.FieldsFunc(input, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func NormalizedString(word string) []string {
	words := splitIntoWords(word)
	out := []string{}
	mp := make(map[string]struct{})

	for _, word := range words {
		normWord, err := snowball.Stem(word, "english", false)
		if err != nil {
			continue
		}
		if _, ok := mp[normWord]; ok {
			continue
		}
		if IsStopWord(normWord) {
			continue
		}

		mp[normWord] = struct{}{}
		out = append(out, normWord)
	}

	return out
}
