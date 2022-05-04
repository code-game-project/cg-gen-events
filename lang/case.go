package lang

import "strings"

func IsSnakeCaseIdentifier(text string) bool {
	runes := []rune(text)
	for i, r := range runes {
		if i == 0 && !(r >= 'a' && r <= 'z' || r == '_') {
			return false
		} else if !(r >= 'a' && r <= 'z' || r == '_' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

func snakeToCamel(text string) string {
	parts := strings.Split(text, "_")
	for i, p := range parts {
		if i > 0 {
			parts[i] = strings.Title(p)
		}
	}
	return strings.Join(parts, "")
}

func snakeToPascal(text string) string {
	text = strings.ReplaceAll(text, "_", " ")
	text = strings.Title(text)
	text = strings.ReplaceAll(text, " ", "")
	return text
}

func snakeToKebab(text string) string {
	return strings.ReplaceAll(text, "-", "")
}

func snakeToOneWord(text string) string {
	return strings.ReplaceAll(text, "_", "")
}
