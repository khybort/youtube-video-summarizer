package textprocessing

import (
	"strings"
	"unicode"
)

type Preprocessor struct {
	MaxLength int
	Language  string
}

func NewPreprocessor(maxLength int, language string) *Preprocessor {
	return &Preprocessor{
		MaxLength: maxLength,
		Language:  language,
	}
}

func (p *Preprocessor) PreprocessForEmbedding(text string) string {
	// 1. Remove excessive whitespace
	text = strings.Join(strings.Fields(text), " ")

	// 2. Normalize unicode (basic)
	text = strings.ToLower(text)

	// 3. Remove URLs (simple regex would be better)
	// For now, just trim

	// 4. Truncate to max length
	if p.MaxLength > 0 && len(text) > p.MaxLength {
		text = text[:p.MaxLength]
		// Try to cut at word boundary
		if lastSpace := strings.LastIndex(text, " "); lastSpace > 0 {
			text = text[:lastSpace]
		}
	}

	// 5. Handle empty strings
	if strings.TrimSpace(text) == "" {
		return ""
	}

	return text
}

func (p *Preprocessor) ChunkText(text string, chunkSize, overlap int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var chunks []string
	var currentChunk []string
	currentSize := 0

	for _, word := range words {
		currentChunk = append(currentChunk, word)
		currentSize++

		if currentSize >= chunkSize {
			chunks = append(chunks, strings.Join(currentChunk, " "))
			// Keep overlap words for next chunk
			if overlap > 0 && len(currentChunk) > overlap {
				currentChunk = currentChunk[len(currentChunk)-overlap:]
				currentSize = overlap
			} else {
				currentChunk = []string{}
				currentSize = 0
			}
		}
	}

	// Add remaining chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}

	return chunks
}

func (p *Preprocessor) CleanText(text string) string {
	// Remove special characters but keep basic punctuation
	var result strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) ||
			r == '.' || r == ',' || r == '!' || r == '?' || r == ':' || r == ';' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

