package interpreter

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/playwright-community/playwright-go"
)

// Interpreter управляет извлечением элементов со страницы
type Interpreter struct {
	page playwright.Page
}

// New создаёт новый Interpreter
func New(page playwright.Page) *Interpreter {
	return &Interpreter{
		page: page,
	}
}

// Snapshot возвращает список интерактивных элементов на странице через axe-core
func (i *Interpreter) Snapshot() ([]Element, error) {
	_, err := i.page.AddScriptTag(playwright.PageAddScriptTagOptions{
		URL: playwright.String("https://cdnjs.cloudflare.com/ajax/libs/axe-core/4.7.2/axe.min.js"),
	})
	if err != nil {
		return nil, fmt.Errorf("AddScriptTag axe-core: %w", err)
	}

	handle, err := i.page.EvaluateHandle(`
		async () => {
			const results = await axe.run(document.body);
			const elements = [];
			results.violations.forEach(v => {
				v.nodes.forEach(n => {
					elements.push({
						id: v.id,
						role: n.html || '(no role)',
						name: n.any && n.any.length > 0 ? n.any[0].relatedNodes[0]?.html || '(no label)' : '(no label)',
						disabled: n.element?.disabled || false
					});
				});
			});
			return elements.slice(0, 100);
		}
	`)
	if err != nil {
		return nil, fmt.Errorf("EvaluateHandle axe-core JS: %w", err)
	}

	jsonValue, err := handle.JSONValue()
	if err != nil {
		return nil, fmt.Errorf("JSONValue axe-core result: %w", err)
	}

	jsonBytes, err := json.Marshal(jsonValue)
	if err != nil {
		return nil, fmt.Errorf("marshal JSONValue: %w", err)
	}
	var elements []Element
	if err = json.Unmarshal(jsonBytes, &elements); err != nil {
		return nil, fmt.Errorf("unmarshal to []Element: %w", err)
	}

	return elements, nil
}

// PrintSnapshot печатает элементы для отладки
func (i *Interpreter) PrintSnapshot() {
	els, err := i.Snapshot()
	if err != nil {
		log.Println("Snapshot error:", err)
		return
	}

	log.Printf("Found %d interactive elements:\n", len(els))
	for _, el := range els {
		log.Printf("[%s] %s (%s) disabled=%v\n", el.ID, el.Name, el.Role, el.Disabled)
	}
}
