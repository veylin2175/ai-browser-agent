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
	// 1. Ждём, пока body появится и будет не пустым
	_, err := i.page.WaitForFunction(`
        () => document.body && document.body.children.length > 0
    `, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		return nil, fmt.Errorf("страница не загрузилась (body не найден): %w", err)
	}

	// 2. Сам evaluate
	resultHandle, err := i.page.EvaluateHandle(`
    () => {
        // Защита от отсутствия body
        if (!document.body) {
            return [];
        }

        const interactiveTags = new Set([
            "A", "BUTTON", "INPUT", "SELECT", "TEXTAREA", "LABEL",
            "SUMMARY", "DETAILS", "[tabindex]:not([tabindex='-1'])"
        ]);

        const interactiveRoles = new Set([
            "button", "link", "checkbox", "radio", "textbox", "searchbox",
            "combobox", "listbox", "menuitem", "tab", "switch"
        ]);

        function isInteractive(el) {
            if (!el || el.nodeType !== Node.ELEMENT_NODE) return false;

            const style = window.getComputedStyle(el);
            if (style.display === "none" || style.visibility === "hidden" || style.opacity === "0") {
                return false;
            }

            const rect = el.getBoundingClientRect();
            if (rect.width <= 4 || rect.height <= 4) return false;

            const tag = el.tagName;
            const role = (el.getAttribute("role") || "").toLowerCase();

            return interactiveTags.has(tag) ||
                   interactiveRoles.has(role) ||
                   (el.hasAttribute("tabindex") && el.tabIndex >= 0);
        }

        function getBestName(el) {
            let name = (
                el.getAttribute("aria-label") ||
                (el.getAttribute("aria-labelledby") && document.getElementById(el.getAttribute("aria-labelledby"))?.textContent?.trim()) ||
                el.placeholder ||
                el.alt ||
                el.title ||
                el.textContent?.trim().replace(/\\s+/g, " ") ||
                el.value ||
                "(без имени)"
            );
            return name.slice(0, 100).replace(/[\n\r]+/g, " ");
        }

        function buildSimpleSelector(el) {
            if (!el) return "";
            if (el.id) return "#" + el.id;

            const parts = [];
            let current = el;
            let depth = 0;

            while (current && current !== document.body && depth < 7) {
                let part = current.tagName.toLowerCase();
                if (current.id) {
                    part += "#" + current.id;
                    parts.unshift(part);
                    break;
                }
                if (current.className && typeof current.className === "string") {
                    const cls = current.className.trim().split(/\\s+/).filter(Boolean);
                    if (cls.length) part += "." + cls.join(".");
                }
                parts.unshift(part);
                current = current.parentElement;
                depth++;
            }
            return parts.join(" > ") || el.tagName.toLowerCase();
        }

        const elements = [];
        let index = 0;

        // TreeWalker — только если body существует
        const walker = document.createTreeWalker(
            document.body,
            NodeFilter.SHOW_ELEMENT,
            { acceptNode: node => NodeFilter.FILTER_ACCEPT }
        );

        while (walker.nextNode() && index < 130) {
            const node = walker.currentNode;
            if (isInteractive(node)) {
                elements.push({
                    index: index++,
                    selector: buildSimpleSelector(node),
                    role: (node.getAttribute("role") || node.tagName.toLowerCase()),
                    name: getBestName(node),
                    disabled: !!node.disabled,
                    visible: true  // уже отфильтровали по rect
                });
            }
        }

        // Fallback: querySelectorAll на самые частые интерактивные элементы
        if (elements.length === 0) {
            const fallback = document.querySelectorAll("a, button, input, select, textarea, [role=button], [role=link], [tabindex]");
            fallback.forEach(el => {
                if (isInteractive(el)) {
                    elements.push({
                        index: index++,
                        selector: buildSimpleSelector(el),
                        role: (el.getAttribute("role") || el.tagName.toLowerCase()),
                        name: getBestName(el),
                        disabled: !!el.disabled,
                        visible: true
                    });
                }
            });
        }

        return elements;
    }`)

	if err != nil {
		return nil, fmt.Errorf("ошибка evaluate: %w", err)
	}

	jsonData, err := resultHandle.JSONValue()
	if err != nil {
		return nil, fmt.Errorf("JSONValue failed: %w", err)
	}

	bytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	var elements []Element
	err = json.Unmarshal(bytes, &elements)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	return elements, nil
}

// PrintSnapshot для отладки
func (i *Interpreter) PrintSnapshot() {
	els, err := i.Snapshot()
	if err != nil {
		log.Println("Snapshot error:", err)
		return
	}

	log.Printf("Found %d interactive elements:\n", len(els))
	for _, el := range els {
		log.Printf("[%d] %s (%s) disabled=%v\n", el.Index, el.Name, el.Role, el.Disabled)
	}
}
