package promts

import (
	"fmt"
	"strings"

	"ai-browser-agent/internal/interpreter"
)

func BuildSnapshotPrompt(elements []interpreter.Element) string {
	var sb strings.Builder
	sb.WriteString("Индекс | Селектор | Роль | Название | Disabled | InViewport\n")
	sb.WriteString("------|----------|------|----------|----------|------------\n")

	for _, el := range elements {
		name := strings.ReplaceAll(el.Name, "\n", " ")
		name = strings.ReplaceAll(name, `"`, `\"`)
		if len(name) > 80 {
			name = name[:77] + "..."
		}

		selector := el.Selector
		if len(selector) > 60 {
			selector = selector[:57] + "..."
		}

		sb.WriteString(fmt.Sprintf(
			"%d | %s | %s | %q | %v | %v\n",
			el.Index,
			selector,
			el.Role,
			name,
			el.Disabled,
			el.InViewport,
		))
	}
	return sb.String()
}
