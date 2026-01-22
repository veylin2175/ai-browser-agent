package promts

import (
	"fmt"
	"strings"

	"ai-browser-agent/internal/interpreter"
)

func BuildSnapshotPrompt(elements []interpreter.Element) string {
	var sb strings.Builder
	sb.WriteString("Индекс | Селектор | Роль | Название | Disabled\n")
	sb.WriteString("------|----------|------|----------|---------\n")

	for _, el := range elements {
		name := strings.ReplaceAll(el.Name, "\n", " ")
		name = strings.ReplaceAll(name, `"`, `\"`)
		if len(name) > 80 {
			name = name[:77] + "..."
		}

		sb.WriteString(fmt.Sprintf(
			"%d | %s | %s | %q | %v\n",
			el.Index,
			el.Selector,
			el.Role,
			name,
			el.Disabled,
		))
	}
	return sb.String()
}
