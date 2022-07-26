package cge

import (
	"fmt"
	"strings"
)

func generateErrorText(message string, lineText []rune, line, columnStart, columnEnd int) string {
	if columnEnd >= len(lineText) {
		lineText = append(lineText, []rune(strings.Repeat(" ", columnEnd-(len(lineText)-1)))...)
	}

	length := len(lineText)
	lineText = []rune(strings.TrimPrefix(strings.TrimPrefix(string(lineText), " "), "\t"))
	columnStart = columnStart - (length - len(lineText))
	columnEnd = columnEnd - (length - len(lineText))

	errorLine := string(lineText[:columnStart])
	errorLine = errorLine + "\x1b[4m\x1b[31m"
	errorLine = errorLine + string(lineText[columnStart:columnEnd])
	errorLine = errorLine + "\x1b[0m"
	errorLine = errorLine + string(lineText[columnEnd:])

	text := fmt.Sprintf("\x1b[2m[%d]  \x1b[0m%s", line+1, errorLine)
	text = fmt.Sprintf("%s%s\n%s\n%s", fmt.Sprintf("[%d:%d] %s\n", line+1, columnStart+1, message), strings.Repeat("-", 30), text, strings.Repeat("-", 30))
	return text
}
