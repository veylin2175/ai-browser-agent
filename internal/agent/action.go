package agent

type ActionType string

const (
	ActionClick    ActionType = "click"
	ActionTypeText ActionType = "type"
	ActionNavigate ActionType = "navigate"
	ActionWait     ActionType = "wait"
	ActionAskUser  ActionType = "ask_user"
)

type Action struct {
	Type      ActionType `json:"type"`
	ElementID string     `json:"element_id,omitempty"`
	Text      string     `json:"text,omitempty"`
	URL       string     `json:"url,omitempty"`
	Reason    string     `json:"reason,omitempty"`
}
