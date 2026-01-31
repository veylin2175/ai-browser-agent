package interpreter

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/playwright-community/playwright-go"
)

type Interpreter struct {
	page playwright.Page
}

func New(page playwright.Page) *Interpreter {
	return &Interpreter{
		page: page,
	}
}

func (i *Interpreter) Snapshot() ([]Element, error) {
	_, err := i.page.WaitForFunction(`
        () => document.body && document.body.children.length > 0
    `, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		return nil, fmt.Errorf("страница не загрузилась (body не найден): %w", err)
	}

	_, err = i.page.WaitForFunction(`
        () => document.readyState === 'complete' &&
              document.body &&
              document.body.children.length > 0
    `, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		log.Printf("Предупреждение: страница не полностью загрузилась за 15с: %v", err)
	}

	resultHandle, err := i.page.EvaluateHandle(`
    () => {
        if (!document.body) {
            return [];
        }

        const interactiveTags = new Set([
            "A", "BUTTON", "INPUT", "SELECT", "TEXTAREA", "LABEL",
            "SUMMARY", "DETAILS"
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
                   (el.hasAttribute("tabindex") && el.tabIndex >= 0) ||
                   el.onclick != null ||
                   style.cursor === "pointer";
        }

        function getBestName(el) {
            let name = (
                el.getAttribute("aria-label") ||
                (el.getAttribute("aria-labelledby") && document.getElementById(el.getAttribute("aria-labelledby"))?.textContent?.trim()) ||
                el.placeholder ||
                el.alt ||
                el.title ||
                el.textContent?.trim().replace(/\s+/g, " ") ||
                el.value ||
                "(без имени)"
            );
            return name.slice(0, 100).replace(/[\n\r]+/g, " ");
        }

        function buildSimpleSelector(el) {
            if (!el) return "";
            
            if (el.id && /^[a-zA-Z][\w-]*$/.test(el.id)) {
                return "#" + CSS.escape(el.id);
            }

            const parts = [];
            let current = el;
            let depth = 0;

            while (current && current !== document.body && depth < 7) {
                let part = current.tagName.toLowerCase();
                
                if (current.id && /^[a-zA-Z][\w-]*$/.test(current.id)) {
                    part += "#" + CSS.escape(current.id);
                    parts.unshift(part);
                    break;
                }
                
                if (current.className) {
                    const classStr = typeof current.className === "string" 
                        ? current.className 
                        : current.className.baseVal || "";
                    
                    if (classStr) {
                        const cls = classStr.trim().split(/\s+/).filter(c => c && /^[a-zA-Z_-]/.test(c));
                        if (cls.length > 0) {
                            const safeCls = cls.slice(0, 2).map(c => CSS.escape(c));
                            part += "." + safeCls.join(".");
                        }
                    }
                }
                
                if (!current.id) {
                    const classStr = typeof current.className === "string" 
                        ? current.className 
                        : (current.className && current.className.baseVal) || "";
                    
                    if (!classStr || !classStr.trim()) {
                        const siblings = Array.from(current.parentElement?.children || []);
                        const index = siblings.indexOf(current);
                        if (index >= 0 && siblings.length > 1) {
                            part += ":nth-child(" + (index + 1) + ")";
                        }
                    }
                }
                
                parts.unshift(part);
                current = current.parentElement;
                depth++;
            }
            
            return parts.join(" > ") || el.tagName.toLowerCase();
        }

        function isInViewport(el) {
            const rect = el.getBoundingClientRect();
            return (
                rect.top >= 0 &&
                rect.left >= 0 &&
                rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
                rect.right <= (window.innerWidth || document.documentElement.clientWidth)
            );
        }

        const elements = [];
        let index = 0;

        const walker = document.createTreeWalker(
            document.body,
            NodeFilter.SHOW_ELEMENT,
            { acceptNode: node => NodeFilter.FILTER_ACCEPT }
        );

        while (walker.nextNode() && index < 150) {
            const node = walker.currentNode;
            if (isInteractive(node)) {
                const style = window.getComputedStyle(node);
                const rect = node.getBoundingClientRect();

                const isVisible = style.display !== "none" &&
                                 style.visibility !== "hidden" &&
                                 style.opacity !== "0" &&
                                 rect.width > 0 && rect.height > 0;

                elements.push({
                    index: index++,
                    selector: buildSimpleSelector(node),
                    role: (node.getAttribute("role") || node.tagName.toLowerCase()),
                    name: getBestName(node),
                    disabled: !!node.disabled,
                    visible: isVisible,
                    isHidden: !isVisible,
                    inViewport: isInViewport(node)
                });
            }
        }

        if (elements.length === 0) {
            const fallback = document.querySelectorAll("a, button, input, select, textarea, [role=button], [role=link], [tabindex]");
            fallback.forEach(el => {
                if (isInteractive(el)) {
                    const style = window.getComputedStyle(el);
                    const rect = el.getBoundingClientRect();

                    const isVisible = style.display !== "none" &&
                                     style.visibility !== "hidden" &&
                                     style.opacity !== "0" &&
                                     rect.width > 0 && rect.height > 0;

                    elements.push({
                        index: index++,
                        selector: buildSimpleSelector(el),
                        role: (el.getAttribute("role") || el.tagName.toLowerCase()),
                        name: getBestName(el),
                        disabled: !!el.disabled,
                        visible: isVisible,
                        isHidden: !isVisible,
                        inViewport: isInViewport(el)
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
