package executor

import (
	"fmt"
	"time"

	"ai-browser-agent/internal/agent"
	"ai-browser-agent/internal/interpreter"

	"github.com/playwright-community/playwright-go"
)

type Executor struct {
	page        playwright.Page
	interpreter *interpreter.Interpreter
}

func New(page playwright.Page, i *interpreter.Interpreter) *Executor {
	return &Executor{
		page:        page,
		interpreter: i,
	}
}

func (e *Executor) Execute(action agent.Action) error {
	switch action.Type {

	case agent.ActionClick:
		return e.click(action.ElementID)

	case agent.ActionTypeText:
		return e.typeText(action.ElementID, action.Text)

	case agent.ActionNavigate:
		_, err := e.page.Goto(action.URL)
		return err

	case agent.ActionWait:
		time.Sleep(2 * time.Second)
		return nil

	case agent.ActionAskUser:
		return fmt.Errorf("USER_INPUT_REQUIRED: %s", action.Reason)

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}
