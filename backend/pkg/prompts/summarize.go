package prompts

import (
	"fmt"
	"strings"
)

func getShortSummaryPrompt(language string) string {
	languageInstruction := getLanguageInstruction(language)
	return `You are a professional content analyst specializing in video summaries. Create a clear, well-structured summary.

**Instructions:**
- Write a concise summary in 2-3 paragraphs
- Focus on main topic, key insights, and conclusions
- Use clear, professional language
- Extract 3-5 key points as bullet points

**Transcript:**
{{.Transcript}}

**Required Format (use Markdown):**
## Summary

[Write your summary here in 2-3 well-structured paragraphs. Each paragraph should focus on a specific aspect: introduction/context, main content, and conclusions.]

## Key Points

- [First key point - be specific and actionable]
- [Second key point - be specific and actionable]
- [Third key point - be specific and actionable]
- [Fourth key point - optional]
- [Fifth key point - optional]

**Important:**
- Use proper Markdown formatting (## for headings, - for bullets)
- Write in a professional, clear, and engaging style
- Ensure the summary is self-contained and informative
- Key points should be concise (one line each) and meaningful

` + languageInstruction
}

func getDetailedSummaryPrompt(language string) string {
	languageInstruction := getLanguageInstruction(language)
	return `You are a professional content analyst specializing in comprehensive video analysis. Create a detailed, well-structured summary.

**Instructions:**
- Write a comprehensive summary in 4-6 paragraphs
- Cover all major topics, insights, and conclusions
- Include context, main discussion points, and actionable takeaways
- Extract 5-8 key takeaways as bullet points
- Identify main topics covered

**Transcript:**
{{.Transcript}}

**Required Format (use Markdown):**
## Summary

[Write your comprehensive summary here in 4-6 well-structured paragraphs. Structure as follows:
- Paragraph 1: Introduction and context
- Paragraph 2-4: Main content and discussion points
- Paragraph 5-6: Conclusions, insights, and implications]

## Key Takeaways

- [First takeaway - be specific and actionable]
- [Second takeaway - be specific and actionable]
- [Third takeaway - be specific and actionable]
- [Fourth takeaway]
- [Fifth takeaway]
- [Additional takeaways as needed]

## Main Topics Covered

- [Topic 1 - brief description]
- [Topic 2 - brief description]
- [Additional topics as needed]

**Important:**
- Use proper Markdown formatting (## for headings, - for bullets)
- Write in a professional, analytical style
- Ensure comprehensive coverage of all important aspects
- Takeaways should be specific, actionable, and valuable

` + languageInstruction
}

func getBulletPointsPrompt(language string) string {
	languageInstruction := getLanguageInstruction(language)
	return `You are a professional content analyst. Extract and organize the most important information from this video transcript.

**Instructions:**
- Extract 5-10 main points
- Include key facts, statistics, and data points
- Capture important quotes or statements
- Organize information logically
- Each point should be clear and self-contained

**Transcript:**
{{.Transcript}}

**Required Format (use Markdown):**
## Main Points

- [First main point - be specific and informative]
- [Second main point - be specific and informative]
- [Third main point - be specific and informative]
- [Continue with additional points...]

## Key Facts & Statistics

- [Fact or statistic 1 - include numbers/data if mentioned]
- [Fact or statistic 2 - include numbers/data if mentioned]
- [Additional facts as applicable]

## Important Quotes

- "[Quote 1 - if any notable statements were made]"
- "[Quote 2 - if any notable statements were made]"

**Important:**
- Use proper Markdown formatting (## for headings, - for bullets)
- Each bullet point should be informative and complete
- Include specific details, numbers, or data when available
- If no quotes are notable, you may omit that section

` + languageInstruction
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

