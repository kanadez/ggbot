package core

import "strings"

func EscapeStringForMarkdown(str string) string {
	str_escaped := strings.Replace(str, "_", "\\_", -1)
	str_escaped = strings.Replace(str_escaped, "*", "\\*", -1)
	str_escaped = strings.Replace(str_escaped, "[", "\\[", -1)
	str_escaped = strings.Replace(str_escaped, "`", "\\`", -1)

	return str_escaped
}
