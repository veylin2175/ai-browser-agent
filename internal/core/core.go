package core

import "fmt"

type ActionType string

const (
	ActionClick    ActionType = "click"
	ActionTypeText ActionType = "type"
	ActionNavigate ActionType = "navigate"
	ActionDone     ActionType = "done"
	ActionPressKey ActionType = "press_key"
)

type Action struct {
	Type   ActionType `json:"type"`
	Target int        `json:"target,omitempty"`
	Text   string     `json:"text,omitempty"`
	URL    string     `json:"url,omitempty"`
	Reason string     `json:"reason"`
	Key    string     `json:"key,omitempty"`
}

func (a Action) String() string {
	switch a.Type {
	case ActionNavigate:
		return fmt.Sprintf("🌐 Перехожу на %s", a.URL)
	case ActionClick:
		return fmt.Sprintf("🛠️ Нажимаю на %d", a.Target)
	case ActionTypeText:
		textSnippet := a.Text
		if len(textSnippet) > 30 {
			textSnippet = textSnippet[:27] + "..."
		}
		return fmt.Sprintf("🛠️ Ввожу \"%s\" в поле %d", textSnippet, a.Target)
	case ActionDone:
		return "Задача выполнена! 🎉"
	default:
		return fmt.Sprintf("%s (неизвестный тип)", a.Type)
	}
}
