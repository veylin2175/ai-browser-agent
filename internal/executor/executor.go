package executor

import (
	"ai-browser-agent/internal/core"
)

type Executor interface {
	Execute(action *core.Action) error
}
