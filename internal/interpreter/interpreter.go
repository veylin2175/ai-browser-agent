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
	handle, err := i.page.EvaluateHandle(`
		async () => {
		  await new Promise(r => setTimeout(r, 100)); // даём DOM стабилизироваться
		
		  const elements = [];
		  const nodes = document.querySelectorAll('a,button,input,textarea,select');
		
		  let idx = 0;
		  for (const el of nodes) {
			const rect = el.getBoundingClientRect();
			if (rect.width === 0 || rect.height === 0) continue;
		
			let selector = '';
			if (el.id) {
			  selector = '#' + el.id;
			} else if (el.name) {
			  selector = el.tagName.toLowerCase() + '[name="' + el.name + '"]';
			} else {
			  const siblings = Array.from(el.parentNode.children)
				.filter(e => e.tagName === el.tagName);
			  const nth = siblings.indexOf(el) + 1;
			  selector = el.tagName.toLowerCase() + ':nth-of-type(' + nth + ')';
			}
		
			elements.push({
			  id: 'e' + idx++,
			  role: el.tagName.toLowerCase(),
			  name: el.innerText || el.value || el.getAttribute('aria-label') || '(no label)',
			  disabled: !!el.disabled,
			  selector: selector
			});
		  }
		
		  return elements;
		}
	`)
	if err != nil {
		return nil, fmt.Errorf("EvaluateHandle snapshot: %w", err)
	}

	value, err := handle.JSONValue()
	if err != nil {
		return nil, fmt.Errorf("JSONValue snapshot: %w", err)
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}

	var elements []Element
	if err := json.Unmarshal(bytes, &elements); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
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
