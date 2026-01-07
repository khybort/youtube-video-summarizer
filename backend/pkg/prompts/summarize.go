package prompts

import (
	"fmt"
	"strings"
)

func getShortSummaryPrompt(language string) string {
	languageInstruction := getLanguageInstruction(language)
	return `You are a helpful assistant that summarizes YouTube video transcripts.

Given the following transcript, provide:
1. A concise summary in 2-3 paragraphs focusing on the main topic, key points, and conclusions.
2. Key points as a bulleted list (3-5 main points).

Transcript:
{{.Transcript}}

Format your response as:
SUMMARY:
[Your summary here in 2-3 paragraphs]

KEY POINTS:
- [Point 1]
- [Point 2]
- [Point 3]

` + languageInstruction
}

func getDetailedSummaryPrompt(language string) string {
	languageInstruction := getLanguageInstruction(language)
	return `You are a helpful assistant that creates detailed summaries of YouTube videos.

Given the following transcript, provide:
1. A comprehensive summary (4-5 paragraphs)
2. Key takeaways (bullet points)
3. Main topics covered
4. Any actionable insights

Transcript:
{{.Transcript}}

Format your response as:
SUMMARY:
[Your comprehensive summary here]

KEY TAKEAWAYS:
- [Point 1]
- [Point 2]
- [Point 3]

TOPICS:
- [Topic 1]
- [Topic 2]

` + languageInstruction
}

func getBulletPointsPrompt(language string) string {
	languageInstruction := getLanguageInstruction(language)
	return `Analyze the following video transcript and extract:
- Main points (5-10 bullet points)
- Key facts or statistics mentioned
- Important quotes or statements

Transcript:
{{.Transcript}}

Format as a bulleted list. ` + languageInstruction
}

// getLanguageInstruction returns the language instruction based on the language setting
func getLanguageInstruction(language string) string {
	if language == "" || language == "auto" {
		return "Provide your response in the same language as the transcript."
	}
	
	// Map common language codes to language names
	languageMap := map[string]string{
		"en": "English",
		"tr": "Turkish",
		"es": "Spanish",
		"fr": "French",
		"de": "German",
		"it": "Italian",
		"pt": "Portuguese",
		"ru": "Russian",
		"ja": "Japanese",
		"ko": "Korean",
		"zh": "Chinese",
		"ar": "Arabic",
		"hi": "Hindi",
		"nl": "Dutch",
		"pl": "Polish",
		"sv": "Swedish",
		"da": "Danish",
		"no": "Norwegian",
		"fi": "Finnish",
	}
	
	langName, ok := languageMap[strings.ToLower(language)]
	if ok {
		return fmt.Sprintf("Provide your response in %s.", langName)
	}
	
	// If language code not found, use it as-is (might be a full language name)
	return fmt.Sprintf("Provide your response in %s.", language)
}

func GetSummaryPrompt(summaryType string, language string) string {
	switch summaryType {
	case "detailed":
		return getDetailedSummaryPrompt(language)
	case "bullet_points":
		return getBulletPointsPrompt(language)
	default:
		return getShortSummaryPrompt(language)
	}
}

