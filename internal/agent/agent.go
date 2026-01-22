package agent

import (
	"ai-browser-agent/internal/agent/promts"
	"fmt"
	"strings"

	"ai-browser-agent/internal/core"
	"ai-browser-agent/internal/interpreter"
	"ai-browser-agent/internal/llm"
)

type Agent struct {
	llm     llm.Client
	i       *interpreter.Interpreter
	History []string
}

func New(llm llm.Client, i *interpreter.Interpreter) *Agent {
	return &Agent{llm: llm, i: i}
}

func (a *Agent) Step(goal string) (*core.Action, error) {
	elements, err := a.i.Snapshot()
	if err != nil {
		return nil, err
	}

	if len(elements) == 0 {
		return nil, fmt.Errorf("no elements")
	}

	historyStr := ""
	if len(a.History) > 0 {
		historyStr = "ПРЕДЫДУЩИЕ ДЕЙСТВИЯ И РЕЗУЛЬТАТЫ (ОБЯЗАТЕЛЬНО УЧТИ!):\\n" + strings.Join(a.History, "\n") + "\n\n"
		historyStr += "НЕ ПОВТОРЯЙ успешные действия из списка выше. Если действие уже сделано успешно — переходи к следующему или завершай.\n"
	}

	userPrompt := fmt.Sprintf(
		"SYSTEM:\n%s\n\n%sGOAL:\n%s\n\nSNAPSHOT:\n%s",
		promts.SystemPrompt,
		historyStr,
		goal,
		promts.BuildSnapshotPrompt(elements),
	)

	action, err := a.llm.NextAction(userPrompt)
	if err != nil {
		return nil, err
	}

	return action, nil
}
