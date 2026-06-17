package transformers

import (
	"regexp"
	"strings"
)

var (
	ttsXMLTagPattern       = regexp.MustCompile(`</?[A-Za-z][A-Za-z0-9:_-]*(?:\s+[^<>]*)?/?>`)
	ttsBracketURLPattern   = regexp.MustCompile(`(?is)[（(][^()（）]{0,300}(?:https?://|www\.)[^()（）]{0,300}[）)]`)
	ttsBareURLPattern      = regexp.MustCompile(`(?i)\b(?:https?://|www\.)[^\s<>"'，。！？；、）)\]}]+`)
	ttsWhitespacePattern   = regexp.MustCompile(`[ \t\r\n]+`)
	ttsPunctuationReplacer = strings.NewReplacer(
		" ，", "，",
		" 。", "。",
		" ？", "？",
		" ！", "！",
		" ：", "：",
		" ；", "；",
		" 、", "、",
		"( )", "",
		"（ ）", "",
	)
)

func normalizeTTSSpokenText(text string) string {
	if text == "" {
		return ""
	}
	text = ttsBracketURLPattern.ReplaceAllString(text, "")
	text = ttsXMLTagPattern.ReplaceAllString(text, "")
	text = ttsBareURLPattern.ReplaceAllString(text, "")
	text = ttsWhitespacePattern.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	text = ttsPunctuationReplacer.Replace(text)
	return strings.TrimSpace(text)
}
