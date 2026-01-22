package llm

import "ai-browser-agent/internal/core"

type Client interface {
	NextAction(prompt string) (*core.Action, error)
}

type DummyClient struct{}

func NewDummy() Client {
	return &DummyClient{}
}

func (d *DummyClient) NextAction(prompt string) (*core.Action, error) {
	return &core.Action{
		Type:   core.ActionClick,
		Target: 0,
	}, nil
}
