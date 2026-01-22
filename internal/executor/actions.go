package executor

import "fmt"

func (e *Executor) click(elementID string) error {
	els, err := e.interpreter.Snapshot()
	if err != nil {
		return err
	}

	for _, el := range els {
		if el.ID == elementID {
			return e.page.Click(el.Selector)
		}
	}

	return fmt.Errorf("element %s not found", elementID)
}

func (e *Executor) typeText(elementID, text string) error {
	els, err := e.interpreter.Snapshot()
	if err != nil {
		return err
	}

	for _, el := range els {
		if el.ID == elementID {
			return e.page.Fill(el.Selector, text)
		}
	}

	return fmt.Errorf("element %s not found", elementID)
}
